package models

type MessagesReply struct {
	MsgToUser         string
	MsgToAssigner     string
	MsgToPrevAssignee string
	AssignerChatID    int64
	UserChatID        int64
	AssigneeChatID    int64
}
