# Build GoLang aplikacije 
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Ovo omogućava Dockeru da kešira zavisnosti
COPY go.mod go.sum ./

RUN go mod download

# Kopira sav ostatak izvornog koda aplikacije 
COPY . .

# CGO_ENABLED=0: Onemogućava CGO, što rezultira statički linkovanim binarnim fajlom
# -a -installsuffix nocgo: Dodatne opcije za statički link
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o blog-service .

# --- FAZA 2: Konačna minimalna runtime slika ---
FROM alpine:latest

WORKDIR /app

# Kopiraj samo kompajlirani binarni fajl (blog-service) iz faze 'builder' u finalnu sliku
COPY --from=builder /app/blog-service .

EXPOSE 8081

# Pokreće prethodno kompajlirani binarni fajl
ENTRYPOINT ["./blog-service"]