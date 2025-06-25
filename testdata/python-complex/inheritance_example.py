#!/usr/bin/env python3
"""
Complex Python inheritance and decorator examples for integration testing.
"""

from abc import ABC, abstractmethod
from typing import Generic, TypeVar, Protocol, List, Dict
from dataclasses import dataclass
import functools
import asyncio

T = TypeVar("T")


class Drawable(Protocol):
    """Protocol for drawable objects."""

    def draw(self) -> str: ...


class Shape(ABC):
    """Abstract base class for shapes."""

    def __init__(self, name: str, color: str = "black"):
        self.name = name
        self.color = color

    @abstractmethod
    def area(self) -> float:
        """Calculate the area of the shape."""
        pass

    @abstractmethod
    def perimeter(self) -> float:
        """Calculate the perimeter of the shape."""
        pass

    def describe(self) -> str:
        return f"{self.color} {self.name} with area {self.area()}"


class Rectangle(Shape):
    """Rectangle implementation."""

    def __init__(self, width: float, height: float, **kwargs):
        super().__init__("rectangle", **kwargs)
        self.width = width
        self.height = height

    def area(self) -> float:
        return self.width * self.height

    def perimeter(self) -> float:
        return 2 * (self.width + self.height)

    def draw(self) -> str:
        return f"Drawing rectangle {self.width}x{self.height}"


class Square(Rectangle):
    """Square is a special rectangle."""

    def __init__(self, side: float, **kwargs):
        super().__init__(side, side, **kwargs)
        self.name = "square"

    def draw(self) -> str:
        return f"Drawing square {self.width}x{self.width}"


class Circle(Shape):
    """Circle implementation."""

    def __init__(self, radius: float, **kwargs):
        super().__init__("circle", **kwargs)
        self.radius = radius

    def area(self) -> float:
        return 3.14159 * self.radius**2

    def perimeter(self) -> float:
        return 2 * 3.14159 * self.radius

    def draw(self) -> str:
        return f"Drawing circle with radius {self.radius}"


@dataclass
class Point:
    """Point in 2D space."""

    x: float
    y: float

    def distance_to(self, other: "Point") -> float:
        return ((self.x - other.x) ** 2 + (self.y - other.y) ** 2) ** 0.5


class Container(Generic[T]):
    """Generic container class."""

    def __init__(self):
        self._items: List[T] = []

    def add(self, item: T) -> None:
        self._items.append(item)

    def remove(self, item: T) -> bool:
        try:
            self._items.remove(item)
            return True
        except ValueError:
            return False

    def get_all(self) -> List[T]:
        return self._items.copy()

    def __len__(self) -> int:
        return len(self._items)

    def __iter__(self):
        return iter(self._items)


def retry(max_attempts: int = 3):
    """Decorator for retrying failed operations."""

    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            last_exception = None
            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    last_exception = e
                    if attempt < max_attempts - 1:
                        continue
            raise last_exception

        return wrapper

    return decorator


def timer(func):
    """Simple timing decorator."""

    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        import time

        start = time.time()
        result = func(*args, **kwargs)
        end = time.time()
        print(f"{func.__name__} took {end - start:.4f}s")
        return result

    return wrapper


class ShapeManager:
    """Manager for shapes with various operations."""

    def __init__(self):
        self.shapes: Container[Shape] = Container()

    def add_shape(self, shape: Shape) -> None:
        self.shapes.add(shape)

    @timer
    def calculate_total_area(self) -> float:
        return sum(shape.area() for shape in self.shapes)

    @retry(max_attempts=3)
    def save_to_file(self, filename: str) -> None:
        # Simulate potential I/O failure
        import random

        if random.random() < 0.3:
            raise IOError("Random I/O failure")

        with open(filename, "w") as f:
            for shape in self.shapes:
                f.write(f"{shape.describe()}\n")

    @property
    def shape_count(self) -> int:
        return len(self.shapes)

    @staticmethod
    def create_default() -> "ShapeManager":
        manager = ShapeManager()
        manager.add_shape(Rectangle(10, 5))
        manager.add_shape(Square(3))
        manager.add_shape(Circle(2))
        return manager

    @classmethod
    def from_shapes(cls, shapes: List[Shape]) -> "ShapeManager":
        manager = cls()
        for shape in shapes:
            manager.add_shape(shape)
        return manager


class AsyncShapeProcessor:
    """Async processor for shapes."""

    def __init__(self, delay: float = 0.1):
        self.delay = delay

    async def process_shape(self, shape: Shape) -> Dict[str, float]:
        await asyncio.sleep(self.delay)
        return {"area": shape.area(), "perimeter": shape.perimeter()}

    async def process_shapes(self, shapes: List[Shape]) -> List[Dict[str, float]]:
        tasks = [self.process_shape(shape) for shape in shapes]
        return await asyncio.gather(*tasks)


def create_test_shapes() -> List[Shape]:
    """Factory function for creating test shapes."""
    return [
        Rectangle(5, 10, color="red"),
        Square(4, color="blue"),
        Circle(3, color="green"),
        Rectangle(2, 8, color="yellow"),
        Square(6, color="purple"),
    ]


async def main():
    """Async main function demonstrating usage."""
    shapes = create_test_shapes()
    manager = ShapeManager.from_shapes(shapes)

    print(f"Created {manager.shape_count} shapes")
    print(f"Total area: {manager.calculate_total_area()}")

    processor = AsyncShapeProcessor()
    results = await processor.process_shapes(shapes)

    for i, result in enumerate(results):
        print(f"Shape {i}: {result}")


if __name__ == "__main__":
    asyncio.run(main())
