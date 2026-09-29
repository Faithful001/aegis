.PHONY: dev dev-backend dev-frontend

dev:
	npx concurrently -n "backend,frontend" -c "cyan,magenta" "make dev-backend" "make dev-frontend"

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && npm run dev