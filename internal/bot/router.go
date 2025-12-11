package bot

import "strings"

type Response struct {
	Text string
}

func Handle(input string, prefix string) Response {
	trimmed := strings.TrimSpace(input)
	after, found := strings.CutPrefix(trimmed, prefix)

	if !found {
		return Response{Text: ""}
	}

	params := strings.Fields(after)

	var text string
	switch params[0] {
	case "d":
		text = "you called deadline with params " + strings.Join(params[1:], ",")
	case "p":
		text = "you called pins with params " + strings.Join(params[1:], ",")
	case "h":
		text = "you called help"
	default:
		text = "unknown command, use .h for help"
	}

	return Response{Text: text}
}
