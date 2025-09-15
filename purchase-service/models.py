from sqlalchemy import Column, Integer, String, Float, ForeignKey
from sqlalchemy.orm import relationship
from .database import Base

class ShoppingCart(Base):
    __tablename__ = "shopping_carts"

    id = Column(Integer, primary_key=True, index=True)
    tourist_id = Column(Integer, unique=True, index=True) # ID korisnika koji je vlasnik korpe
    total_price = Column(Float, default=0.0)

    # Definišemo vezu: jedna korpa može imati više stavki
    items = relationship("OrderItem", back_populates="cart")

class OrderItem(Base):
    __tablename__ = "order_items"

    id = Column(Integer, primary_key=True, index=True)
    tour_id = Column(Integer, index=True)
    tour_name = Column(String)
    price = Column(Float)
    
    cart_id = Column(Integer, ForeignKey("shopping_carts.id"))
    
    # Definišemo vezu nazad ka korpi
    cart = relationship("ShoppingCart", back_populates="items")

class TourPurchaseToken(Base):
    __tablename__ = "purchase_tokens"

    id = Column(Integer, primary_key=True, index=True)
    token = Column(String, unique=True, index=True) # Jedinstveni token za kupljenu turu
    tour_id = Column(Integer, index=True)
    tourist_id = Column(Integer, index=True)