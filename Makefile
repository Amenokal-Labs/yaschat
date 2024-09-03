.PHONY: start-frontend start-backend

start-frontend:
	cd frontend && npm run dev

start-backend:
	cd backend && go run cmd/messaging-app/main.go
