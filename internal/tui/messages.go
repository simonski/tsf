package tui

// authSuccessMsg is sent when authentication succeeds
type authSuccessMsg struct{}

// changeViewMsg is sent to switch views
type changeViewMsg struct {
	view ViewType
	data interface{}
}

// errorMsg is sent when an error occurs
type errorMsg struct {
	err error
}

// dataLoadedMsg is sent when data is loaded
type dataLoadedMsg struct {
	data interface{}
}
