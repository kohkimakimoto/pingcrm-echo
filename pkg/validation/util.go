package validation

import (
	"fmt"
)

func Message(defaultMsg string, customMsgAndArgs ...interface{}) string {
	if len(customMsgAndArgs) == 0 || customMsgAndArgs == nil {
		return defaultMsg
	}
	if len(customMsgAndArgs) == 1 {
		msg := customMsgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}

	if len(customMsgAndArgs) > 1 {
		return fmt.Sprintf(customMsgAndArgs[0].(string), customMsgAndArgs[1:]...)
	}
	return defaultMsg
}
