package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"github.com/pdbogen/mapbot/common/db"
	"github.com/pdbogen/mapbot/hub"
	"github.com/pdbogen/mapbot/model/user"
	"github.com/pdbogen/mapbot/model/workflow"
	"io"
	"net/http"
	"strings"
)

func writeResponse(rw http.ResponseWriter, msg string) {
	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	body, err := json.Marshal(slack.Msg{
		Text: msg,
	})
	if err != nil {
		log.Errorf("marshalling JSON: %s", err)
		rw.Write([]byte(`{"text": "an error occurred"}`))
		return
	}
	rw.Write(body)
}

// upon receiving an action, we need to pass it to the corresponding workflow with the appropriate state name, opaque data, and choice.
// thus the action's ID will need to let us obtain the workflow name, state name, and opaque data. the choice will come from the action callback itself.
// the response func may return an error, whic we need to send to the user.
// if the response doesn't report an error, we'll call the challenge for the new state; which will give back a WorkflowMessage.
func (s *SlackUi) Action(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if err := req.ParseForm(); err != nil {
		writeResponse(rw, "request could not be parsed")
		log.Errorf("error parsing form: %s", err)
		return
	}

	payloads, ok := req.Form["payload"]
	if !ok || len(payloads) == 0 {
		writeResponse(rw, "request had no payload")
		log.Errorf("no payloads in request")
		return
	}
	var payload *slack.AttachmentActionCallback

	if err := json.Unmarshal([]byte(payloads[0]), &payload); err != nil || payload == nil {
		writeResponse(rw, "error parsing payload")
		log.Errorf("error unmarshalling JSON payload: %s", err)
		return
	}

	if s.verificationToken != payload.Token {
		writeResponse(rw, "forbidden")
		log.Errorf("received token %q that does not match slack verification token", payload.Token)
		return
	}

	var team *Team
	for _, t := range s.Teams {
		if t.Info.ID == payload.Team.ID {
			team = t
		}
	}

	if team == nil {
		writeResponse(rw, "your team is not recognized. mapbot may need to be reinstalled.")
		log.Errorf("team %q received in action not found", payload.Team.ID)
		return
	}
	team.Action(payload, rw, req)
}

func (t *Team) Action(payload *slack.AttachmentActionCallback, rw http.ResponseWriter, req *http.Request) {
	userObj, err := user.Get(db.Instance, user.Id(payload.User.ID))
	if err != nil {
		writeResponse(rw, "could not retrieve user")
		log.Errorf("could not retrieve user %q in action", payload.User.ID)
		return
	}

	writeResponse(rw, "Okay, hold on...")

	t.hub.Publish(&hub.Command{
		User:    userObj,
		From:    fmt.Sprintf("internal:updateAction:slack:%s:%s", t.Info.ID, payload.ResponseURL),
		Context: nil,
		Payload: payload.Actions[0].Value,
		Type:    "user:workflow:respond",
	})
}

func (t *Team) updateAction(h *hub.Hub, c *hub.Command) {
	comps := strings.Split(string(c.Type), ":")
	if len(comps) < 5 {
		log.Errorf("%s: received but cannot process command %s", t.Info.ID, c.Type)
		return
	}
	responseUrl := strings.Join(comps[4:], ":")

	body := &bytes.Buffer{}
	enc := json.NewEncoder(body)
	req, err := http.NewRequest("POST", responseUrl, body)
	if err != nil {
		log.Errorf("creating request: %s", err)
		return
	}
	req.Header.Add("content-type", "application/json")

	switch msg := c.Payload.(type) {
	case *workflow.WorkflowMessage:
		enc.Encode(t.renderWorkflowMessage(msg))
	case string:
		enc.Encode(slack.Msg{Text: msg})
	default:
		log.Errorf("uh, no clue how to handle a %T payload", c.Payload)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("POSTing to %q: %s", responseUrl, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		log.Errorf("non-2XX POSTing to %s: %s", responseUrl, res.Status)
		body := &bytes.Buffer{}
		io.Copy(body, res.Body)
		log.Errorf("body: %q", body.String())
		return
	}
}