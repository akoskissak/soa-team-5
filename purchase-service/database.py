import os
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.declarative import declarative_base

# Učitavamo URL baze iz .env fajla
# Format: "postgresql://user:password@host:port/dbname"
# U našem slučaju: "postgresql://postgres:super@purchasedb:5432/purchase_db"
DATABASE_URL = os.getenv("PURCHASE_DATABASE_URL", "postgresql://postgres:super@localhost:5434/purchase_db")

# Kreiramo SQLAlchemy "engine" koji će upravljati konekcijama
engine = create_engine(DATABASE_URL)

# Kreiramo klasu za sesiju, svaka instanca ove klase će biti jedna sesija sa bazom
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# Ovo je bazna klasa koju će naši SQLAlchemy modeli naslediti
Base = declarative_base()