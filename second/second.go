package second

import (
    "fmt"
    "net/http"
    "time"

    tea "github.com/charmbracelet/bubbletea"
)

const url = "https://charm.sh/"

type Model struct {
    status int
    err    error
}

func checkServer() tea.Msg {

    // Create an HTTP client and make a GET request.
    c := &http.Client{Timeout: 10 * time.Second}
    res, err := c.Get(url)

    if err != nil {
        // There was an error making our request. Wrap the error we received
        // in a message and return it.
        return errMsg{err}
    }
    // We received a response from the server. Return the HTTP status code
    // as a message.
    return statusMsg(res.StatusCode)
}

type statusMsg int

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

func (m Model) Init() (tea.Cmd) {
    return checkServer
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case statusMsg:
        // The server returned a status message. Save it to our model. Also
        // tell the Bubble Tea runtime we want to exit because we have nothing
        // else to do. We'll still be able to render a final view with our
        // status message.
        m.status = int(msg)
        return m, tea.Quit

    case errMsg:
        // There was an error. Note it in the model. And tell the runtime
        // we're done and want to quit.
        m.err = msg
        return m, tea.Quit

    case tea.KeyMsg:
        // Ctrl+c exits. Even with short running programs it's good to have
        // a quit key, just in case your logic is off. Users will be very
        // annoyed if they can't exit.
        if msg.Type == tea.KeyCtrlC {
            return m, tea.Quit
        }
    }

    // If we happen to get any other messages, don't do anything.
    return m, nil
}

func (m Model) View() string {
    // If there's an error, print it out and don't do anything else.
    if m.err != nil {
        return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
    }

    // Tell the user we're doing something.
    s := fmt.Sprintf("Checking %s ... ", url)

    // When the server responds with a status, add it to the current line.
    if m.status > 0 {
        s += fmt.Sprintf("%d %s!", m.status, http.StatusText(m.status))
    }

    // Send off whatever we came up with above for rendering.
    return "\n" + s + "\n\n"
}
