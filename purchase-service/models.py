from sqlalchemy import Column, Integer, String, Float, ForeignKey, DateTime
from sqlalchemy.orm import relationship
from sqlalchemy_utils import UUIDType
from datetime import datetime

from database import Base

class ShoppingCart(Base):
    __tablename__ = "shopping_carts"

    id = Column(Integer, primary_key=True, index=True)
    tourist_id = Column(String, index=True) 
    total_price = Column(Float, default=0.0)

    items = relationship("OrderItem", back_populates="cart")

class OrderItem(Base):
    __tablename__ = "order_items"

    id = Column(Integer, primary_key=True, index=True)
    tour_id = Column(UUIDType(binary=False), index=True)
    tour_name = Column(String)
    price = Column(Float)
    
    cart_id = Column(Integer, ForeignKey("shopping_carts.id"))
    
    cart = relationship("ShoppingCart", back_populates="items")

class TourPurchaseToken(Base):
    __tablename__ = "purchase_tokens"

    id = Column(Integer, primary_key=True, index=True)
    token = Column(String, unique=True, index=True)
    tour_id = Column(UUIDType(binary=False), index=True)
    tourist_id = Column(String, index=True) 
    tour_name = Column(String)
    price = Column(Float)
    created_at = Column(DateTime(timezone=True), nullable=False)