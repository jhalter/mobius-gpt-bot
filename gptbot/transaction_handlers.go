package gptbot

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/jhalter/mobius/hotline"
	"github.com/sashabaranov/go-openai"
	"regexp"
	"strings"
	"time"
)

func (b *Bot) HandleToolCallFunc(toolCallType string, handleFunc toolCallHandleFunc) {
	b.toolCallHandlers[toolCallType] = handleFunc
}

// HandleNotifyDeleteUser removes users from the list of logged-in users on exit.
func (b *Bot) HandleNotifyDeleteUser(ctx context.Context, c *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	exitUser := t.GetField(hotline.FieldUserID).Data

	var newUserList []hotline.User
	for _, u := range c.UserList {
		if !bytes.Equal(exitUser, u.ID) {
			newUserList = append(newUserList, u)
		}
	}

	c.UserList = newUserList

	return res, err
}

func (b *Bot) HandleClientGetUserNameList(_ context.Context, c *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	var users []hotline.User
	for _, field := range t.Fields {
		// The Hotline protocol docs say that ClientGetUserNameList should only return fieldUsernameWithInfo (300)
		// fields, but shxd sneaks in fieldChatSubject (115) so it's important to filter explicitly for the expected
		// field type.  Probably a good idea to do everywhere.
		if bytes.Equal(field.ID, []byte{0x01, 0x2c}) {
			u, err := hotline.ReadUser(field.Data)
			if err != nil {
				return res, err
			}
			users = append(users, *u)
		}
	}
	c.UserList = users

	return res, err
}

func (b *Bot) HandleKeepAlive(_ context.Context, _ *hotline.Client, _ *hotline.Transaction) ([]hotline.Transaction, error) {
	return []hotline.Transaction{}, nil
}

// HandleInviteToChat responds to private chat invitations by accepting the invite.
func (b *Bot) HandleInviteToChat(_ context.Context, _ *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	res = append(
		res,
		*hotline.NewTransaction(
			hotline.TranJoinChat,
			nil,
			hotline.NewField(hotline.FieldChatID, t.GetField(hotline.FieldChatID).Data),
		),
	)

	return res, err
}

// HandleServerMsg reponds to direct messages from users.
func (b *Bot) HandleServerMsg(ctx context.Context, _ *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	msg := strings.ReplaceAll(string(t.GetField(hotline.FieldData).Data), "\r", "\n")
	hlUser := string(t.GetField(hotline.FieldUserName).Data)

	if len(t.GetField(hotline.FieldUserID).Data) != 2 {
		return res, errors.New("invalid request")
	}
	userID := binary.BigEndian.Uint16(t.GetField(hotline.FieldUserID).Data)

	b.HotlineClient.Logger.InfoContext(ctx, "Received private message", "msg", msg, "hlUser", hlUser)

	if _, ok := b.PMThreads[userID]; !ok {
		pubChatThread, err := b.OpenAPIClient.CreateThread(ctx, openai.ThreadRequest{})
		if err != nil {
			return res, err
		}
		b.PMThreads[userID] = pubChatThread
	}

	_, err = b.OpenAPIClient.CreateMessage(ctx, b.PMThreads[userID].ID, openai.MessageRequest{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})
	if err != nil {
		return res, err
	}

	run, err := b.RunLoop(ctx, b.PMThreads[userID].ID)
	if err != nil {
		return res, err
	}

	updatedThread, err := b.OpenAPIClient.ListMessage(ctx, run.ThreadID, nil, nil, nil, nil)
	if err != nil {
		return res, err
	}

	replyMsg := strings.ReplaceAll(updatedThread.Messages[0].Content[0].Text.Value, "\n", "\r")

	reply := hotline.NewTransaction(
		hotline.TranSendInstantMsg,
		nil,
		hotline.NewField(hotline.FieldData, []byte(replyMsg)),
		hotline.NewField(hotline.FieldUserID, t.GetField(hotline.FieldUserID).Data),
	)
	res = append(
		res,
		*reply,
	)

	return res, err
}

// chatMsgRegex matches public chat messages that are addressed to the bot user.
const chatMsgRegex = "(?P<User>\\w*):  (?P<Msg>.*)"

func (b *Bot) HandleClientChatMsg(ctx context.Context, c *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	r := regexp.MustCompile(chatMsgRegex)
	matches := r.FindStringSubmatch(string(t.GetField(hotline.FieldData).Data))
	if len(matches) != 3 {
		return res, errors.New("invalid chat message")
	}
	user := matches[1]
	msg := matches[2]

	var chatInt int
	if len(t.GetField(hotline.FieldChatID).Data) > 0 {
		chatInt = int(binary.BigEndian.Uint32(t.GetField(hotline.FieldChatID).Data))
	}

	// If message came from the bot, ignore it to avoid an infinite self-reply loop
	if user == b.Config.Name {
		return res, nil
	}

	if chatInt == 0 {
		b.ChatLogs = append(b.ChatLogs, PubChatLog{
			Username: user,
			Message:  msg,
			Time:     time.Now(),
		})

		// If the incoming message is in public chat, check if it is addressed to the bot user
		br := regexp.MustCompile(fmt.Sprintf(`(?i):\s+%s[:,\s]+(.*$)`, b.Config.Name))
		if !br.Match(t.GetField(hotline.FieldData).Data) {
			return res, nil
		}
	}

	b.HotlineClient.Logger.Debug(
		"got chat message",
		"user", user,
		"msg", msg,
		"chatID", chatInt,
	)

	var chatThread openai.Thread
	if _, ok := b.Threads[chatInt]; !ok {
		chatThread, err = b.OpenAPIClient.CreateThread(ctx, openai.ThreadRequest{})
		if err != nil {
			return res, fmt.Errorf("openAI createThread error: %w", err)
		}
		b.Threads[chatInt] = chatThread
	}

	if time.Now().Add(-contextTTL).After(b.lastInteraction) {
		_, err = b.OpenAPIClient.DeleteThread(ctx, b.Threads[chatInt].ID)

		b.Threads[chatInt], err = b.OpenAPIClient.CreateThread(ctx, openai.ThreadRequest{})
		if err != nil {
			return res, err
		}

	}

	_, err = b.OpenAPIClient.CreateMessage(ctx, b.Threads[chatInt].ID, openai.MessageRequest{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})
	if err != nil {
		return res, fmt.Errorf("openAI createMessage error: %w", err)
	}

	run, err := b.RunLoop(ctx, b.Threads[chatInt].ID)
	if err != nil {
		return res, err
	}

	limit := 1
	updatedThread, err := b.OpenAPIClient.ListMessage(ctx, run.ThreadID, &limit, nil, nil, nil)
	if err != nil {
		return res, fmt.Errorf("openAI listMessage error: %w", err)
	}

	fields := []hotline.Field{
		hotline.NewField(hotline.FieldData, []byte(strings.ReplaceAll(updatedThread.Messages[0].Content[0].Text.Value, "\n", "\r"))),
	}

	if chatInt != 0 {
		fields = append(fields,
			hotline.NewField(hotline.FieldChatID, t.GetField(hotline.FieldChatID).Data),
		)
	}

	res = append(res,
		*hotline.NewTransaction(hotline.TranChatSend, nil,
			fields...,
		),
	)

	// update last interaction time
	b.lastInteractionMUX.Lock()
	defer b.lastInteractionMUX.Unlock()
	b.lastInteraction = time.Now()

	return res, err
}

const newsSeparator = "__________________________________________________________"

func (b *Bot) TranGetClientInfoText(ctx context.Context, c *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	r := regexp.MustCompile(`\A.*\rAccount:\s+(\w.*)\rAddress:\s+(\d.*):`)
	matches := r.FindStringSubmatch(string(t.GetField(hotline.FieldData).Data))

	if len(matches) != 3 {
		return res, errors.New("unable to get user info: possibly missing Can Get User Info permission.  user greeting disabled.")
	}

	account := matches[1]
	ipAddr := matches[2]

	b.usersMUX.Lock()
	defer b.usersMUX.Unlock()
	visitor, ok := b.Users[ipAddr]
	if !ok {
		b.Users[ipAddr] = NewUser(account)
		visitor = b.Users[ipAddr]
	}

	if !visitor.greetingLimiter.Allow() {
		c.Logger.Info("greeting rate limit exceeded; skipping welcome message", "user", visitor.ipAddr)

		return res, err
	}

	greetThread, err := b.OpenAPIClient.CreateThread(ctx, openai.ThreadRequest{})
	if err != nil {
		return res, err
	}

	_, err = b.OpenAPIClient.CreateMessage(ctx, greetThread.ID, openai.MessageRequest{
		Role:    openai.ChatMessageRoleUser,
		Content: fmt.Sprintf(b.Config.Greeting, string(t.GetField(hotline.FieldUserName).Data)),
	})
	if err != nil {
		return res, fmt.Errorf("openAI createMessage error: %w", err)
	}

	run, err := b.RunLoop(ctx, b.Threads[0].ID)
	if err != nil {
		return res, err
	}

	updatedThread, err := b.OpenAPIClient.ListMessage(ctx, run.ThreadID, nil, nil, nil, nil)
	if err != nil {
		return res, err
	}

	b.Visitors = append(b.Visitors,
		Visitor{
			Username: string(t.GetField(hotline.FieldUserName).Data),
			Time:     time.Now(),
		},
	)

	c.Logger.Info("Sent new user greeting", "content", updatedThread.Messages[0].Content[0].Text.Value)

	res = append(res,
		*hotline.NewTransaction(hotline.TranChatSend, nil,
			hotline.NewField(hotline.FieldData, []byte(strings.ReplaceAll(updatedThread.Messages[0].Content[0].Text.Value, "\n", "\r"))),
		),
	)

	return res, err
}

func (b *Bot) TranGetMsgs(ctx context.Context, c *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	newsText := string(t.GetField(hotline.FieldData).Data)
	newsText = strings.ReplaceAll(newsText, "\r", "\n")

	b.FlatNews = strings.Split(newsText, newsSeparator)

	return res, err
}

func (b *Bot) TranNotifyChangeUser(_ context.Context, _ *hotline.Client, t *hotline.Transaction) (res []hotline.Transaction, err error) {
	newUser := hotline.User{
		ID:    t.GetField(hotline.FieldUserID).Data,
		Name:  string(t.GetField(hotline.FieldUserName).Data),
		Icon:  t.GetField(hotline.FieldUserIconID).Data,
		Flags: t.GetField(hotline.FieldUserFlags).Data,
	}

	if string(t.GetField(hotline.FieldUserName).Data) == "" {
		return res, nil
	}

	// Check to see if this is transaction was triggered by a new visitor to the server, or a status change to an
	// existing user.  In the latter case we don't need to do anything.
	for i := 0; i < len(b.HotlineClient.UserList); i++ {
		if bytes.Equal(newUser.ID, b.HotlineClient.UserList[i].ID) {
			return res, nil
		}
	}

	b.HotlineClient.UserList = append(b.HotlineClient.UserList, newUser)

	// If we're configured to greet users, send a request to get user info so that we can check the user IP address for
	// rate limiting purposes.
	if b.Config.GreetUsers {
		res = append(res,
			*hotline.NewTransaction(hotline.TranGetClientInfoText, nil,
				hotline.NewField(hotline.FieldUserID, t.GetField(hotline.FieldUserID).Data),
			),
		)
	}
	return res, err
}
