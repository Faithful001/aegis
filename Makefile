.PHONY: dev dev-backend dev-frontend

dev:
	npx concurrently -k -n "backend,frontend" -c "cyan,magenta" "cd backend && air" "cd frontend && npm run dev"

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && npm run dev