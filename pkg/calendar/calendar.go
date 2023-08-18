package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// addEventToGoogleCalendar adds an event to the Google Calendar using the provided duty date, duty type, and employee name.
// It reads the credentials from the path_to_credentials.json file and creates a new event with the provided information.
// The event is inserted into the primary calendar and the link to the created event is logged.
func addEventToGoogleCalendar(googleClientSecretPath string, googleTokenPath string, dutyDate time.Time, dutyType string, employeeName string) {
	ctx := context.Background()

	srv, err := createCalendarService(&ctx, googleClientSecretPath, googleTokenPath)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v\n", err)
	}

	event := &calendar.Event{
		Summary:     "Дежурство: " + dutyType,
		Description: "Дежурный: " + employeeName,
		Start: &calendar.EventDateTime{
			DateTime: dutyDate.Format(time.RFC3339),
			TimeZone: "Europe/Moscow",
		},
		End: &calendar.EventDateTime{
			DateTime: dutyDate.Add(24 * time.Hour).Format(time.RFC3339),
			TimeZone: "Europe/Moscow",
		},
	}

	calendarId := "primary" // используйте "primary" для основного календаря
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	log.Printf("Event created: %s\n", event.HtmlLink)
}

// функция deleteEventFromGoogleCalendar удаляет событие из Google Calendar по ссылке на событие.
func deleteEventFromGoogleCalendar(googleClientSecretPath string, googleTokenPath string, eventLink string) {
	ctx := context.Background()

	srv, err := createCalendarService(&ctx, googleClientSecretPath, googleTokenPath)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v\n", err)
	}

	err = srv.Events.Delete("primary", eventLink).Do()
	if err != nil {
		log.Fatalf("Unable to delete event. %v\n", err)
	}
	log.Printf("Event deleted: %s\n", eventLink)
}

func createCalendarService(ctx *context.Context, googleClientSecretPath string, googleTokenPath string) (*calendar.Service, error) {
	b, err := os.ReadFile(googleClientSecretPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл credentials.json: %v", err)
	}

	// Если вы изменяете эти области, удалите файл token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("не удалось проанализировать файл credentials.json в конфиг: %v", err)
	}

	client := getClient(ctx, config, googleTokenPath)

	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать клиент Calendar service: %v", err)
	}

	return srv, nil
}

// getClient returns an authenticated HTTP client using the provided oauth2.Config and context.Context.
// It first attempts to retrieve a token from a file specified by cacheFile, and if that fails, it retrieves a new token from the web.
// The retrieved token is then saved to the cacheFile for future use.
func getClient(ctx *context.Context, config *oauth2.Config, googleTokenPath string) *http.Client {
	tok, err := tokenFromFile(googleTokenPath)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(googleTokenPath, tok)
	}
	return config.Client(*ctx, tok)
}

// getTokenFromWeb retrieves an OAuth2 token from the user by prompting them to visit the authorization URL and enter the received code.
// It takes a pointer to an oauth2.Config struct as input and returns a pointer to an oauth2.Token struct.
// If an error occurs during the process, it logs a fatal error and exits the program.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Перейдите по следующей ссылке в вашем браузере и введите полученный код:\n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Невозможно прочитать код аутентификации: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Невозможно получить токен по этому коду: %v", err)
	}
	return tok
}

func tokenFromFile(googleTokenPath string) (*oauth2.Token, error) {
	f, err := os.Open(googleTokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(googleTokenPath string, token *oauth2.Token) {
	fmt.Printf("Сохранение креденциалов в: %s\n", googleTokenPath)
	f, err := os.OpenFile(googleTokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Невозможно сохранить креденциалы: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
