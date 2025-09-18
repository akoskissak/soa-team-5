from typing import List
import uuid
import httpx
from fastapi import FastAPI, Depends, HTTPException, status
from sqlalchemy.orm import Session
from sqlalchemy import or_

# Ispravni importi
import models, schemas
from database import SessionLocal, engine

models.Base.metadata.create_all(bind=engine)

app = FastAPI()

# --- Dependency ---
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
def create_shopping_cart_for_tourist(tourist_id: str, db: Session = Depends(get_db)):
    db_cart = db.query(models.ShoppingCart).filter(models.ShoppingCart.tourist_id == tourist_id).first()
    if db_cart:
        return db_cart
    
    new_cart = models.ShoppingCart(tourist_id=tourist_id)
    db.add(new_cart)
    db.commit()
    db.refresh(new_cart)
    return new_cart

@app.post("/api/shopping-cart/{tourist_id}/items", response_model=schemas.ShoppingCart)
async def add_item_to_cart(tourist_id: str, item: schemas.OrderItemCreate, db: Session = Depends(get_db)):
    db_cart = db.query(models.ShoppingCart).filter(models.ShoppingCart.tourist_id == tourist_id).first()
    if not db_cart:
        db_cart = models.ShoppingCart(tourist_id=tourist_id)
        db.add(db_cart)
        db.commit()
        db.refresh(db_cart)

    existing_item = db.query(models.OrderItem).filter(
        models.OrderItem.cart_id == db_cart.id,
        models.OrderItem.tour_id == item.tour_id
    ).first()
    
    if existing_item:
        raise HTTPException(status_code=400, detail="This tour is already in the cart")

    db_item = models.OrderItem(**item.model_dump(), cart_id=db_cart.id)
    db.add(db_item)
    
    db_cart.total_price += item.price
    
    db.commit()
    db.refresh(db_cart)
    
    return db_cart


@app.get("/api/shopping-cart/{tourist_id}", response_model=schemas.ShoppingCart)
def get_shopping_cart_by_tourist_id(tourist_id: str, db: Session = Depends(get_db)):
    db_cart = db.query(models.ShoppingCart).filter(models.ShoppingCart.tourist_id == tourist_id).first()
    if not db_cart:
        raise HTTPException(status_code=404, detail="Shopping cart not found")
    return db_cart

@app.delete("/api/shopping-cart/{tourist_id}/items/{tour_id}", status_code=status.HTTP_204_NO_CONTENT)
def remove_item_from_cart(tourist_id: str, tour_id: uuid.UUID, db: Session = Depends(get_db)):
    db_cart = db.query(models.ShoppingCart).filter(models.ShoppingCart.tourist_id == tourist_id).first()
    if not db_cart:
        raise HTTPException(status_code=404, detail="Shopping cart not found")

    db_item = db.query(models.OrderItem).filter(
        models.OrderItem.cart_id == db_cart.id,
        models.OrderItem.tour_id == tour_id
    ).first()

    if not db_item:
        raise HTTPException(status_code=404, detail="Item not found in cart")

    db_cart.total_price -= db_item.price
    db.delete(db_item)
    db.commit()
    return {"message": "Item successfully removed"}

# --- API Rute za Checkout ---

@app.post("/api/shopping-cart/{tourist_id}/checkout", response_model=List[schemas.TourPurchaseToken])
def checkout(tourist_id: str, db: Session = Depends(get_db)):
    db_cart = db.query(models.ShoppingCart).filter(models.ShoppingCart.tourist_id == tourist_id).first()
    if not db_cart or not db_cart.items:
        raise HTTPException(status_code=404, detail="Shopping cart is empty or not found")

    tokens = []
    
    for item in db_cart.items:
        new_token = models.TourPurchaseToken(
            tour_id=item.tour_id,
            tourist_id=tourist_id,
            tour_name=item.tour_name,
            price = item.price,
            token=str(uuid.uuid4()) 
        )
        db.add(new_token)
        tokens.append(new_token)

    for item in db_cart.items:
        db.delete(item)
    
    db_cart.total_price = 0.0 
    
    db.commit()
    
    for token in tokens:
        db.refresh(token)
    
    return tokens

# --- API Rute za kupljene ture ---

@app.get("/api/tourist/{tourist_id}/purchases", response_model=List[schemas.TourPurchaseToken])
def get_tourist_purchases(tourist_id: str, db: Session = Depends(get_db)):
    purchases = db.query(models.TourPurchaseToken).filter(
        models.TourPurchaseToken.tourist_id == tourist_id
    ).all()
    
    if not purchases:
        raise HTTPException(status_code=404, detail="No purchases found for this tourist")
    
    return purchases

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8088, reload=True)