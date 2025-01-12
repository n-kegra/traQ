package viewer

import (
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// State 閲覧状態
type State int

const (
	// StateNone バックグランド表示中
	StateNone State = iota
	// StateMonitoring メッセージ表示中
	StateMonitoring
	// StateEditing メッセージ入力中
	StateEditing
)

// String string表記にします
func (s State) String() string {
	return viewStateStrings[s]
}

// MarshalJSON encoding/json.Marshaler 実装
func (s State) MarshalJSON() ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(s.String())
}

// StateFromString stringからviewer.Stateに変換します
func StateFromString(s string) State {
	return stringViewStates[strings.ToLower(s)]
}

var (
	viewStateStrings = map[State]string{
		StateNone:       "none",
		StateEditing:    "editing",
		StateMonitoring: "monitoring",
	}
	stringViewStates = map[string]State{}
)

func init() {
	// 転置マップ生成
	for v, k := range viewStateStrings {
		stringViewStates[k] = v
	}
}
