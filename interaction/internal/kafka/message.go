package kafka

import "encoding/json"

const (
	TopicLikeAction = "like-action"
	
	ActionLike   = "like"
	ActionUnlike = "unlike"
)

type LikeMessage struct {
	Id        string `json:"id"`
	TargetType int32  `json:"target_type"`
	TargetId  int64  `json:"target_id"`
	UserId    int64  `json:"user_id"`
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
}

func (m *LikeMessage) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func DecodeLikeMessage(data []byte) (*LikeMessage, error) {
	var msg LikeMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
