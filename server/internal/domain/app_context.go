package domain

import "context"

type RequestSourceType string

const RequestSourceTypeHTTP RequestSourceType = "http"

type HTTPRequestSource struct {
	Endpoint string
}

type RequestSource struct {
	Type RequestSourceType
	HTTP *HTTPRequestSource
}

type AppContext struct {
	ctx           context.Context
	user          *User
	requestSource RequestSource
	flowID        string
}

func (a AppContext) Context() context.Context {
	return a.ctx
}

func (a *AppContext) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a AppContext) User() *User {
	return a.user
}

func (a *AppContext) SetUser(user *User) {
	a.user = user
}

func (a AppContext) RequestSource() RequestSource {
	return a.requestSource
}

func (a *AppContext) SetRequestSource(requestSource RequestSource) {
	a.requestSource = requestSource
}

func (a AppContext) FlowID() string {
	return a.flowID
}

func (a *AppContext) SetFlowID(flowID string) {
	a.flowID = flowID
}
