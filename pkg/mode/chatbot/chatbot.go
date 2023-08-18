package chatbot

import (
	"bytes"
	"context"
	"database/sql"
	"dev-support-schedule/pkg/handlers"
	"dev-support-schedule/pkg/models"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	db                     *sql.DB
	webhookURL             string
	domain                 string
	googleClientSecretPath string
	googleTokenPath        string
}

type BotMessage struct {
	Type       string      `json:"type"`
	ID         int64       `json:"id"`
	Event      string      `json:"event"`
	EntityType string      `json:"entity_type"`
	EntityID   int64       `json:"entity_id"`
	ChatID     int64       `json:"chat_id"`
	Content    string      `json:"content"`
	UserID     int64       `json:"user_id"`
	CreatedAt  time.Time   `json:"created_at"`
	Thread     interface{} `json:"thread"` // или *ThreadStruct если у вас есть отдельная структура для "thread"
}

func NewServer(db *sql.DB, webhookURL, domain, googleClientSecretPath, googleTokenPath string) *Server {
	return &Server{
		db:                     db,
		webhookURL:             webhookURL,
		domain:                 domain,
		googleClientSecretPath: googleClientSecretPath,
		googleTokenPath:        googleTokenPath,
	}
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	message := BotMessage{}
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(message.Content, "/schedule") {
		command := s.parseCommand(message.Content)
		responseMessage := s.handleCommand(command)
		s.sendResponseToChat(responseMessage)
	}
}

func (s *Server) parseCommand(commandStr string) models.Command {
	var cmd models.Command
	cmd.Statuses = make(map[string]string)

	if strings.HasPrefix(commandStr, "/schedule") {
		parts := strings.Split(commandStr, " ")
		if len(parts) < 2 {
			cmd.Type = "generate"
			return cmd
		}

		switch parts[1] {
		case "help":
			cmd.Type = "help"
		case "generate":
			cmd.Type = "generate"
		case "save":
			cmd.Type = "save"
		case "team":
			cmd.Type = "team"
		case "status":
			cmd.Type = "status"
			for i := 2; i < len(parts); i++ {
				if strings.HasPrefix(parts[i], "@") {
					nickname := strings.TrimPrefix(parts[i], "@")
					if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "@") {
						status := strings.TrimSuffix(parts[i+1], ",")
						cmd.Statuses[nickname] = status
					}
				}
			}
		case "add":
			cmd.Type = "add"
			cmd.Nickname = strings.TrimPrefix(parts[2], "@")
			cmd.FullName = strings.Join(parts[3:], " ")
		default:
			cmd.Type = "help"
		}
	}

	return cmd
}

func (s *Server) sendResponseToChat(message string) {
	payload := map[string]string{"message": message}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return
	}
	_, err = http.Post(s.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return
	}
}

func (s *Server) Start(useHTTPS bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)

	var server *http.Server

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if useHTTPS {
		// Настройка autocert
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,               // Принимаем условия Let's Encrypt
			HostPolicy: autocert.HostWhitelist(s.domain), // Указываем домен, для которого будем получать сертификат
			Cache:      autocert.DirCache("letsencrypt"), // Папка для хранения сертификатов
		}

		server = &http.Server{
			Addr:      ":443",
			TLSConfig: certManager.TLSConfig(),
			Handler:   mux,
		}

		go func() {
			// Пустые строки, т.к. certManager сам найдет и подставит нужные сертификаты
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()
		log.Printf("https server on %s started", s.domain)
	} else {
		server = &http.Server{
			Addr:    ":8080",
			Handler: mux,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()
		log.Println("http server started")
	}

	<-done
	log.Print("Server Stopped")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}

func (s *Server) handleCommand(command models.Command) string {
	ch := handlers.NewCommandHandler(s.db)
	switch command.Type {
	case "generate":
		return ch.GenerateSchedule(false)
	case "save":
		return ch.GenerateSchedule(true)
	case "team":
		return ch.AllEmployees()
	case "status":
		return ch.UpdateEmployeeStatus(command)
	case "add":
		return ch.AddEmployee(command.Nickname, command.FullName)
	case "help":
		return `
	**Привет! Я 🤖 бот, который умеет формировать расписание дежурств.**
	Я понимаю следующие команды:
	*/schedule * - показать расписание на следующую неделю
	*/schedule save* - показать и сохранить расписание на следующую неделю
	*/schedule team* - команда и её статусы
	*/schedule status @nickname stat, @nickname2 stat, ...* - обновить статусы сотрудников (stat: available, sick, vacation, fired)
	*/schedule add @nickname ФИО* - добавить нового сотрудника
	*/schedule help* - выводит это сообщение
		`
	default:
		return "Неизвестная команда"
	}
}
