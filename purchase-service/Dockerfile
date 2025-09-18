# Koristimo zvanični, lagani Python image
FROM python:3.11-slim

# Postavljamo radni direktorijum unutar kontejnera
WORKDIR /app

# Kopiramo fajl sa zavisnostima i instaliramo ih
# Ovo radimo odvojeno da bi iskoristili Docker-ov layer caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Kopiramo ostatak koda aplikacije
COPY . .

# Komanda koja će se izvršiti kada se kontejner pokrene
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8088"]