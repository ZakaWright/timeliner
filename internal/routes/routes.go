package routes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"timeliner/internal/app"
	"timeliner/internal/broadcaster"
	"timeliner/internal/models"
	"timeliner/web/components"

	//"github.com/a-h/templ"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	//"golang.org/x/text/message"
)

type Handler struct {
	App *app.App
	Broadcaster *broadcaster.Broadcaster
}

func NewRouter(app *app.App) http.Handler {
	// create handler
	b := broadcaster.NewBroadcaster()
	go b.Listen()
	h := &Handler{App: app, Broadcaster: b}
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// Routes
	router.Group(func(router chi.Router) {
		router.Use(jwtauth.Verifier(app.Auth.JwtAuth))
		router.Use(jwtauth.Authenticator)

		//router.Get("/", h.Index)
		router.Get("/incidents", h.Incidents)
		//router.Get("/incidents/new", h.NewIncident)
		router.Route("/incidents", func(router chi.Router) {
			router.Get("/", h.Incidents)

			router.Route("/new", func(router chi.Router) {
				router.Get("/", h.NewIncident)
				router.Post("/", h.MakeNewIncident)
			})
		})

		router.Route("/incident", func(router chi.Router) {
			router.Route("/{id}", func(router chi.Router) {
				router.Get("/", h.GetIncident)
				router.Post("/", h.PostIncident)

				router.Post("/close", h.CloseIncident)
				router.Post("/reopen", h.ReopenIncident)

				router.Get("/timeline", h.GetTimeline)
				router.Get("/timeline-events", h.GetTimelineEvents)
				router.Get("/overview", h.GetIncidentOverview)

				router.Route("/endpoints", func(router chi.Router) {
					router.Get("/", h.GetIncidentEndpoints)
					router.Get("/list", h.GetIncidentEndpointsList)

					router.Get("/new-inline", h.InlineNewEndpoint)
					router.Post("/new-inline", h.PostNewEndpointInline)

					router.Get("/new", h.GetNewEndpoint)
					router.Post("/new", h.PostNewEndpoint)
				})
				router.Route("/events", func(router chi.Router) {
					router.Get("/", h.GetIncidentEvents)
					router.Get("/stream", h.EventStream)

					router.Route("/new", func(router chi.Router) {
						router.Get("/", h.GetNewEvent)
						router.Post("/", h.PostNewEvent)
					})
				})

				router.Route("/event", func(router chi.Router) {
					router.Get("/{event_id}/details", h.GetEventDetails)
					router.Post("/{event_id}/new-comment", h.PostNewComment)
				})
			})
		})
	})

	// Public routes
	router.Group(func(router chi.Router) {
		router.Use(jwtauth.Verifier(app.Auth.JwtAuth))
		router.Get("/", h.Index)
	})
	// returns an empty div to clear an element
	router.Get("/empty", h.Empty)
	router.Get("/test", h.Test)

	router.Route("/login", func(router chi.Router) {
		router.Get("/", h.Login)
		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			app.Auth.LoginUser(w, r, app.Models.Users)
		})
	})

	router.Route("/register", func(router chi.Router) {
		router.Get("/", h.RegisterUser)
		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Calling Register User")
			app.Auth.RegisterUser(w, r, app.Models.Users)
		})

	})

	router.Route("/logout", func(router chi.Router) {
		router.Get("/", h.Logout)
		router.Post("/", app.Auth.LogOutUser)
	})

	router.Get("/user/{id}", h.GetUser)
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	return router
}

func (h *Handler) validate_user(r *http.Request) *models.User {
	user, err := h.App.Auth.GetUser(r, h.App.Models.Users)
	if err != nil {
		fmt.Printf("Validation error: %v", err)
		return nil
	}
	return user
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	component := components.Index(h.validate_user(r))
	component.Render(r.Context(), w)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	//component := components.Login(h.validate_user(r))
	component := components.Login()
	component.Render(r.Context(), w)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	component := components.RegisterUser(h.validate_user(r))
	component.Render(r.Context(), w)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	component := components.LogOut()
	component.Render(r.Context(), w)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := h.App.Models.Users.GetByID(intID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	component := components.User(*user)
	component.Render(r.Context(), w)

}

func (h *Handler) Incidents(w http.ResponseWriter, r *http.Request) {
	open_incidents, err := h.App.Models.Incident.GetOpenIncidents()
	if err != nil {
		//http.Error(w, "Failed to fetch open incidents", http.StatusInternalServerError)
		open_incidents = []*models.Incident{}
	}
	closed_incidents, err := h.App.Models.Incident.GetClosedIncidents()
	if err != nil {
		//http.Error(w, "Failed to fetch closed incidents", http.StatusInternalServerError)
		closed_incidents = []*models.Incident{}
	}
	component := components.Incidents(h.validate_user(r), open_incidents, closed_incidents)
	component.Render(r.Context(), w)
}

func (h *Handler) NewIncident(w http.ResponseWriter, r *http.Request) {
	component := components.NewIncident(h.validate_user(r))
	component.Render(r.Context(), w)
}

func (h *Handler) MakeNewIncident(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// read values
	name := r.PostForm.Get("incident-name")
	caseNumber := r.PostForm.Get("case-number")
	description := r.PostForm.Get("description")
	status := r.PostForm.Get("status")
	if status == "" {
		status = "open"
	} else {
		status = "closed"
	}
	// get user
	user := h.validate_user(r)

	var incident models.Incident
	// create the object
	incident.Name = name
	incident.CaseNumber = caseNumber
	incident.Description = description
	incident.Status = status
	incident.CreatedBy = user.ID
	// post it
	incident_id, err := h.App.Models.Incident.Insert(incident.Name, incident.Description, incident.CaseNumber, incident.Status, incident.CreatedBy)
	if err != nil {
		if strings.Contains(err.Error(), "failed to create incident") {
			fmt.Printf("Erorr: %v\n", err)
			//w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div id="incident-toast-error" class="toast show" role="alert" aria-live="assertive" aria-atomic="true">
                <div class="toast-header">
                    <strong>Erorr Creating Incident</strong>      
                	<button class="btn-close" type="button" data-bs-dismiss="toast" aria-label="close"></button>
				</div>
            </div>
            <div class="toast-body">Failed to create incident
            </div>`)
			return
		}
		if (strings.Contains(err.Error(), "user is inactive")) || (strings.Contains(err.Error(), "no user found")) {
			//w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div id="incident-toast-error" class="toast show" role="alert" aria-live="assertive" aria-atomic="true">
                <div class="toast-header">
                    <strong>Erorr Creating Incident</strong>
                	<button class="btn-close" type="button" data-bs-dismiss="toast" aria-label="close"></button>
				</div>
            </div>
            <div class="toast-body">Invalid User
            </div>`)
			return
		}
	}
	// redirect to /incidents/{id}
	//http.Redirect(w, r, "/incidents/"+strconv.FormatInt(incident_id, 10), http.StatusSeeOther)
	w.Header().Set("HX-Redirect", "/incident/"+strconv.FormatInt(incident_id, 10))
	w.WriteHeader(http.StatusOK)

}

func (h *Handler) GetIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	incident, err := h.App.Models.Incident.GetByID(intID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Incident does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Error retrieving incident", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return
	}
	user := h.validate_user(r)
	incident_user, err := h.App.Models.Users.GetByID(incident.CreatedBy)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusBadRequest)
	}
	component := components.Incident(user, incident, incident_user)
	component.Render(r.Context(), w)
}

func (h *Handler) PostIncident(w http.ResponseWriter, r *http.Request) {
	//TODO
}
func (h *Handler) GetIncidentEndpoints(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return
	}
	endpoints, err := h.App.Models.Endpoints.GetByIncidentID(intID)
	if err != nil {
		http.Error(w, "Issue retreiving endpoints", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return

	}
	component := components.Endpoints(intID, endpoints)
	component.Render(r.Context(), w)
}
func (h *Handler) GetIncidentEndpointsList(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
	}
	endpoints, err := h.App.Models.Endpoints.GetNamesByIncidentID(intID)
	if err != nil {
		http.Error(w, "Issue retreiving endpoints", http.StatusBadRequest)
	}
	component := components.EndpointsList(endpoints)
	component.Render(r.Context(), w)
}
func (h *Handler) GetIncidentEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}
	events, err := h.App.Models.Events.GetEventsForIncident(intID)
	if err != nil {
		http.Error(w, "Problem receiving events", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return
	}

	component := components.Events(intID, events)
	component.Render(r.Context(), w)

}
func (h *Handler) GetNewEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}
	iocTypes, err := h.App.Models.Events.GetIOCTypes()
	if err != nil {
		http.Error(w, "Error receiving IOC Types", http.StatusBadRequest)
	}
	mitreTactics, err := h.App.Models.Events.GetMitreTactics()
	if err != nil {
		http.Error(w, "Error receiving MITRE Tactics", http.StatusBadRequest)
	}
	component := components.NewEvent(intID, iocTypes, mitreTactics)
	component.Render(r.Context(), w)
}
func (h *Handler) GetNewEndpoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	component := components.NewEndpoint(intID)
	component.Render(r.Context(), w)
}
func (h *Handler) InlineNewEndpoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	component := components.InlineNewEndpoint(intID)
	component.Render(r.Context(), w)
}
func (h *Handler) PostNewEndpoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Endpoint ID", http.StatusBadRequest)
	}
	r.ParseForm()
	// read values
	name := r.PostForm.Get("endpoint-name")
	os := r.PostForm.Get("endpoint-os")
	ip := r.PostForm.Get("endpoint-ip")
	mac := r.PostForm.Get("endpoint-mac")
	last_seen := r.PostForm.Get("endpoint-last-seen")

	// get user
	//user := h.validate_user(r)

	var endpoint models.Endpoint
	// create the object
	endpoint.Name = name
	endpoint.OS = os
	endpoint.IP = ip
	endpoint.Mac = mac
	endpoint.IncidentID = intID

	// Parse last_seen string to *time.Time
	var parsedLastSeen *time.Time
	if last_seen != "" {
		t, err := time.Parse("2006-01-02T15:04:05", last_seen)
		if err == nil {
			parsedLastSeen = &t
		}
	}
	endpoint.Last_Seen = parsedLastSeen

	// post it
	_, err = h.App.Models.Endpoints.Insert(&endpoint)
	//incident_id, err := h.App.Models.Incident.Insert(incident.Name, incident.Description, incident.CaseNumber, incident.Status, incident.CreatedBy)
	if err != nil {
		if strings.Contains(err.Error(), "failed to create endpoint") {
			fmt.Printf("Erorr: %v\n", err)
			//w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div id="incident-toast-error" class="toast show" role="alert" aria-live="assertive" aria-atomic="true">
                <div class="toast-header">
                    <strong>Erorr Creating Incident</strong>      
                	<button class="btn-close" type="button" data-bs-dismiss="toast" aria-label="close"></button>
				</div>
            </div>
            <div class="toast-body">Failed to create incident
            </div>`)
			return
		}
	}

	endpoints, err := h.App.Models.Endpoints.GetByIncidentID(intID)
	if err != nil {
		http.Error(w, "Issue retreiving endpoints", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return

	}
	component := components.Endpoints(intID, endpoints)
	component.Render(r.Context(), w)
}
func (h *Handler) PostNewEndpointInline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Endpoint ID", http.StatusBadRequest)
	}
	r.ParseForm()
	// read values
	name := r.PostForm.Get("endpoint-name")
	os := r.PostForm.Get("endpoint-os")
	ip := r.PostForm.Get("endpoint-ip")
	mac := r.PostForm.Get("endpoint-mac")
	last_seen := r.PostForm.Get("endpoint-last-seen")

	// get user
	//user := h.validate_user(r)

	var endpoint models.Endpoint
	// create the object
	endpoint.Name = name
	endpoint.OS = os
	endpoint.IP = ip
	endpoint.Mac = mac
	endpoint.IncidentID = intID

	// Parse last_seen string to *time.Time
	var parsedLastSeen *time.Time
	if last_seen != "" {
		t, err := time.Parse("2006-01-02 15:04:05", last_seen)
		if err == nil {
			parsedLastSeen = &t
		}
	}
	endpoint.Last_Seen = parsedLastSeen

	// post it
	_, err = h.App.Models.Endpoints.Insert(&endpoint)
	//incident_id, err := h.App.Models.Incident.Insert(incident.Name, incident.Description, incident.CaseNumber, incident.Status, incident.CreatedBy)
	if err != nil {
		if strings.Contains(err.Error(), "failed to create endpoint") {
			fmt.Printf("Erorr: %v\n", err)
			//w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div id="incident-toast-error" class="toast show" role="alert" aria-live="assertive" aria-atomic="true">
                <div class="toast-header">
                    <strong>Erorr Creating Incident</strong>      
                	<button class="btn-close" type="button" data-bs-dismiss="toast" aria-label="close"></button>
				</div>
            </div>
            <div class="toast-body">Failed to create incident
            </div>`)
			return
		}
	}

	component := components.Empty()
	component.Render(r.Context(), w)
}
func (h *Handler) PostNewEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Endpoint ID", http.StatusBadRequest)
	}
	r.ParseForm()
	// read values
	//name := r.PostForm.Get("event-name")
	eventTime := r.PostForm.Get("event-time")
	//timeZone := r.PostForm.Get("event-time-tz")
	event_type := r.PostForm.Get("event-type")
	endpoint := r.PostForm.Get("event-endpoint")
	description := r.PostForm.Get("event-description")
	iocTypes := r.PostForm["ioc-type"]
	iocValues := r.PostForm["ioc-value"]
	tactic := r.PostForm.Get("event-tactic")

	// get user
	user := h.validate_user(r)

	var timestamp pgtype.Timestamptz
	if eventTime != "" {

		t, err := time.Parse("2006-01-02T15:04", eventTime)
		if err != nil {
			fmt.Printf("Error parsing time: %v\n", err)
		}
		timestamp = pgtype.Timestamptz{
			Time: t,
			Status: pgtype.Present,
			InfinityModifier: pgtype.None,
		}
	}

	var parsedEndpoint int64
	if endpoint != "" {
		e, err := strconv.ParseInt(endpoint, 10, 64)
		if err == nil {
			parsedEndpoint = e
		}
	}

	var event models.Event
	// create the object
	event.Incident = intID
	//event.EventTime = parsedTime
	event.EventTime = timestamp
	event.EventType = event_type
	event.Description = description
	event.CreatedBy = user.ID
	event.Endpoint = parsedEndpoint
	event.MitreTactic = tactic

	// post it
	event_id, err := h.App.Models.Events.Insert(&event)
	//incident_id, err := h.App.Models.Incident.Insert(incident.Name, incident.Description, incident.CaseNumber, incident.Status, incident.CreatedBy)
	if event_id == -1 {
		fmt.Println("ERROR IN INSERT")

	}

	for i := range len(iocTypes) {
		_, err := h.App.Models.Events.InsertIOC(event_id, user.ID, iocTypes[i], iocValues[i])
		if err != nil {
			fmt.Printf("Error in IOC Insert %v", err)
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), "failed to create") {
			//w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div id="incident-toast-error" class="toast show" role="alert" aria-live="assertive" aria-atomic="true">
                <div class="toast-header">
                    <strong>Erorr Creating Event</strong>      
                	<button class="btn-close" type="button" data-bs-dismiss="toast" aria-label="close"></button>
				</div>
            </div>
            <div class="toast-body">Failed to create event
            </div>`)
			return
		}
	}
	// broadcast the event creation with SSE
	message := broadcaster.Message{IncidentID: intID, Message: fmt.Sprintf("NewEvent;incident_id:%d;event_id:%d", intID, event_id)}
	h.Broadcaster.Broadcaster <-message
	h.GetIncidentEvents(w, r)
}
func (h *Handler) Empty(w http.ResponseWriter, r *http.Request) {

	component := components.Empty()
	component.Render(r.Context(), w)
}
func (h *Handler) CloseIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	err = h.App.Models.Incident.Close(intID)
	if err != nil {
		http.Error(w, "Could not close incident", http.StatusBadRequest)
	}
	incident, err := h.App.Models.Incident.GetByID(intID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Incident does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Error retrieving incident", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return
	}
	//user := h.validate_user(r)
	incident_user, err := h.App.Models.Users.GetByID(incident.CreatedBy)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusBadRequest)
	}
	component := components.IncidentInner(incident, incident_user)
	component.Render(r.Context(), w)
}
func (h *Handler) ReopenIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	err = h.App.Models.Incident.Reopen(intID)
	if err != nil {
		http.Error(w, "Could not close incident", http.StatusBadRequest)
	}
	incident, err := h.App.Models.Incident.GetByID(intID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Incident does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Error retrieving incident", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return
	}
	//user := h.validate_user(r)
	incident_user, err := h.App.Models.Users.GetByID(incident.CreatedBy)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusBadRequest)
	}
	component := components.IncidentInner(incident, incident_user)
	component.Render(r.Context(), w)
}
func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	incident, err := h.App.Models.Incident.GetByID(intID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Incident does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Error retrieving incident", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return
	}
	//user := h.validate_user(r)
	incident_user, err := h.App.Models.Users.GetByID(incident.CreatedBy)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusBadRequest)
	}
	component := components.Timeline(intID, incident_user)
	component.Render(r.Context(), w)
}
func (h *Handler) GetIncidentOverview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	incident, err := h.App.Models.Incident.GetByID(intID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Incident does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Error retrieving incident", http.StatusBadRequest)
		fmt.Printf("Error: %v", err)
		return
	}
	//user := h.validate_user(r)
	incident_user, err := h.App.Models.Users.GetByID(incident.CreatedBy)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusBadRequest)
	}
	component := components.IncidentInner(incident, incident_user)
	component.Render(r.Context(), w)
}
func (h *Handler) GetTimelineEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	events, err := h.App.Models.Events.GetEventDetailsForIncident(intID)
	if err != nil {
		http.Error(w, "Error retreiving events", http.StatusBadRequest)
	}

	component := components.TimelineEvent(intID, events)
	component.Render(r.Context(), w)
}
func (h *Handler) GetEventDetails(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	event_id := chi.URLParam(r, "event_id")
	intEvent_id, err := strconv.ParseInt(event_id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
	}

	eventDetails, err := h.App.Models.Events.GetEventDetails(intEvent_id)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
	}

	component := components.EventDetails(intID, eventDetails)
	component.Render(r.Context(), w)

}
func (h *Handler) PostNewComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	incident_id, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
	}

	id = chi.URLParam(r, "event_id")
	event_id, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
	}
	// get comment
	r.ParseForm()
	newComment := r.PostForm.Get("new-comment")

	// get user
	user := h.validate_user(r)

	err = h.App.Models.Events.AddComment(event_id, user.ID, newComment)
	if err != nil {
		http.Error(w, "Problem adding comment", http.StatusBadRequest)
	}

	event_details, err := h.App.Models.Events.GetEventDetails(event_id)
	if err != nil {
		http.Error(w, "Problem getting event details", http.StatusBadRequest)
	}
	component := components.Comments(incident_id, event_details)
	component.Render(r.Context(), w)
}
func (h *Handler) Test(w http.ResponseWriter, r *http.Request) {
	component := components.Test()
	component.Render(r.Context(), w)
}

func (h *Handler) EventStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	id := chi.URLParam(r, "id")
	incident_id, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
	}
	// set headers to allow all origins.
	// using a tutorial from https://medium.com/@rian.eka.cahya/server-sent-event-sse-with-go-10592d9c2aa1
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// establish new connection channel
	messageChannel := make(chan string)
	client := broadcaster.Client{IncidentID: incident_id, Channel: messageChannel}
	h.Broadcaster.RegisterClient <-client
	defer func() {
		h.Broadcaster.UnregisterClient <-client
	}()
	closedConnection := r.Context().Done()
	go func() {
		<-closedConnection
		h.Broadcaster.UnregisterClient <-client
	}()
	// wait for channel events
	for {
		message := <-messageChannel
		if strings.HasPrefix(message, "NewEvent") {
			// logic for pushing events 
			incident_id_str := strings.TrimPrefix(message, "NewEvent;incident_id:")
			incident_id_str = strings.Split(incident_id_str, ";")[0]
			int_incident_id, err := strconv.ParseInt(incident_id_str, 10, 64)
			if err != nil {
				http.Error(w, "Invalid incident ID", http.StatusBadRequest)
				return
			}
			events, err := h.App.Models.Events.GetEventsForIncident(int_incident_id)
			if err != nil {
				http.Error(w, "Problem receiving events", http.StatusBadRequest)
				fmt.Printf("Error: %v", err)
				return
			}

			component := components.Events(int_incident_id, events)
			html, err := templ.ToGoHTML(context.Background(), component)
			if err != nil {
				// TODO
			}
			fmt.Fprintf(w, "event: NewEvent\n\n")
			fmt.Fprintf(w, "data: %v\n\n", html)
			w.(http.Flusher).Flush()
		}
		flusher.Flush()
	}
}
