from fastapi import FastAPI, Depends, HTTPException
from sqlalchemy.orm import Session

# Importujemo sve što smo napravili u drugim fajlovima
from . import models, schemas
from .database import SessionLocal, engine

# Ova komanda kreira sve tabele u bazi koje smo definisali u models.py
# (ako već ne postoje)
models.Base.metadata.create_all(bind=engine)

app = FastAPI()

# --- Dependency ---
# Ova funkcija obezbeđuje sesiju sa bazom za svaku API rutu.
# FastAPI-jev "Depends" sistem će je pozvati za svaki zahtev.
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

@app.get("/")
def read_root():
    return {"status": "Purchase service is running"}

# --- API Rute ---

@app.post("/api/shopping-cart/{tourist_id}", response_model=schemas.ShoppingCart)
def create_shopping_cart_for_tourist(tourist_id: int, db: Session = Depends(get_db)):
    """
    Kreira (ili vraća postojeću) korpu za datog turistu.
    """
    db_cart = db.query(models.ShoppingCart).filter(models.ShoppingCart.tourist_id == tourist_id).first()
    if db_cart:
        return db_cart
    
    new_cart = models.ShoppingCart(tourist_id=tourist_id)
    db.add(new_cart)
    db.commit()
    db.refresh(new_cart)
    return new_cart

@app.post("/api/shopping-cart/{tourist_id}/items", response_model=schemas.ShoppingCart)
def add_item_to_cart(tourist_id: int, item: schemas.OrderItemCreate, db: Session = Depends(get_db)):
    """
    Dodaje novu stavku (turu) u korpu turiste.
    """
    db_cart = db.query(models.ShoppingCart).filter(models.ShoppingCart.tourist_id == tourist_id).first()
    if not db_cart:
        raise HTTPException(status_code=404, detail="Shopping cart not found for this tourist")

    # Kreiramo novu stavku
    db_item = models.OrderItem(**item.dict(), cart_id=db_cart.id)
    db.add(db_item)
    
    # Ažuriramo ukupnu cenu korpe
    db_cart.total_price += item.price
    
    db.commit()
    db.refresh(db_cart)
    
    return db_cart