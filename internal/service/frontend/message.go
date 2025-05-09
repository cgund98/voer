package frontend

import (
	"fmt"
	"net/http"

	"github.com/ggicci/httpin"
	"google.golang.org/protobuf/proto"

	"github.com/cgund98/voer/internal/entity/db"
	"github.com/cgund98/voer/internal/infra/logging"
	msgComponents "github.com/cgund98/voer/internal/ui/components/message"
)

const pageSize int = 10

type ListMessagesInput struct {
	Page   int    `in:"query=page"`
	Search string `in:"query=search"`
}

// HandleListMessages handles the list messages request
func (s *Service) HandleListMessages(w http.ResponseWriter, r *http.Request) {
	// Parse inputs
	input := r.Context().Value(httpin.Input).(*ListMessagesInput)

	// Fetch messages
	limit := pageSize
	offset := (input.Page - 1) * limit

	messages, err := db.ListMessages(s.db, limit, offset, input.Search)
	if err != nil {
		logging.Logger.Error("Failed to list messages", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := db.CountMessages(s.db, input.Search)
	if err != nil {
		logging.Logger.Error("Failed to count messages", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set hx-trigger header
	w.Header().Set("HX-Trigger", fmt.Sprintf("{\"message-count\": %d}", count))

	// Calculate next page
	nextPage := proto.Int32(int32(input.Page + 1))
	if len(messages) < int(pageSize) {
		nextPage = nil
	}

	// Format messages
	cardInputs := make([]msgComponents.MessageCardInput, len(messages))
	for i, message := range messages {
		msgInput := msgComponents.MessageCardInput{
			Title:   message.Name,
			Package: message.Package.PackageName,
		}
		if message.LatestVersion != nil {
			msgInput.Version = message.LatestVersion.Version
			msgInput.UpdatedAt = message.LatestVersion.UpdatedAt
		}
		cardInputs[i] = msgInput
	}

	// Render component
	component := msgComponents.CardsList(nextPage, cardInputs)
	err = component.Render(r.Context(), w)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error rendering cards list: %v", err))
	}
}
