#!/usr/bin/env python3
"""Models module for testing Python class parsing."""

from typing import List, Dict, Optional, Any, Protocol
from datetime import datetime
from dataclasses import dataclass
from abc import ABC, abstractmethod


# Constants for testing
MAX_USERS = 100
DEFAULT_TIMEOUT = 30.0
SERVICE_VERSION = "1.0.0"
DEBUG_MODE = True

# Module-level variables
global_config: Optional["Config"] = None
user_count = 0
service_registry: Dict[str, Any] = {}
is_initialized = False


@dataclass
class Address:
    """Represents a user's address."""

    street: str
    city: str
    state: str
    zip_code: str
    country: str = "USA"

    def __str__(self) -> str:
        """Return formatted address string."""
        return (
            f"{self.street}, {self.city}, {self.state} {self.zip_code}, {self.country}"
        )

    def validate(self) -> bool:
        """Validate the address fields."""
        if not self.street.strip():
            return False
        if not self.city.strip():
            return False
        return True


class Config:
    """Application configuration class."""

    def __init__(self, database_url: str = "localhost:5432", port: int = 8080):
        self.database_url = database_url
        self.port = port
        self.log_level = "info"
        self.features: Dict[str, bool] = {}

    def set_feature(self, feature: str, enabled: bool) -> None:
        """Enable or disable a feature."""
        self.features[feature] = enabled

    def is_feature_enabled(self, feature: str) -> bool:
        """Check if a feature is enabled."""
        return self.features.get(feature, False)

    @classmethod
    def create_default(cls) -> "Config":
        """Create a default configuration."""
        config = cls()
        config.set_feature("logging", True)
        config.set_feature("metrics", False)
        return config


class User:
    """Represents a user in the system."""

    def __init__(self, user_id: int, name: str, email: str):
        """Initialize a new User."""
        self.id = user_id
        self.name = name
        self.email = email
        self.is_active = True
        self.created_at = datetime.now()
        self.profile: Optional["Profile"] = None

    def __str__(self) -> str:
        """Return string representation of user."""
        return f"User(id={self.id}, name='{self.name}')"

    def activate(self) -> None:
        """Activate the user."""
        self.is_active = True

    def deactivate(self) -> None:
        """Deactivate the user."""
        self.is_active = False

    def set_profile(self, profile: "Profile") -> None:
        """Set the user's profile."""
        self.profile = profile
        profile.user_id = self.id


class Profile:
    """User profile with extended information."""

    def __init__(self, user_id: int):
        """Initialize a new Profile."""
        self.user_id = user_id
        self.bio = ""
        self.address: Optional[Address] = None
        self.skills: List[str] = []
        self.created_at = datetime.now()
        self.updated_at = datetime.now()

    def add_skill(self, skill: str) -> None:
        """Add a skill to the profile."""
        if skill not in self.skills:
            self.skills.append(skill)
            self.updated_at = datetime.now()

    def remove_skill(self, skill: str) -> None:
        """Remove a skill from the profile."""
        if skill in self.skills:
            self.skills.remove(skill)
            self.updated_at = datetime.now()

    def get_skill_count(self) -> int:
        """Return the number of skills."""
        return len(self.skills)

    def set_address(self, address: Address) -> None:
        """Set the profile address."""
        self.address = address
        self.updated_at = datetime.now()


# Abstract base classes and interfaces
class Repository(ABC):
    """Abstract repository interface."""

    @abstractmethod
    def save(self, entity: Any) -> bool:
        """Save an entity."""
        pass

    @abstractmethod
    def find_by_id(self, entity_id: int) -> Optional[Any]:
        """Find entity by ID."""
        pass

    @abstractmethod
    def find_all(self) -> List[Any]:
        """Find all entities."""
        pass

    @abstractmethod
    def delete(self, entity_id: int) -> bool:
        """Delete entity by ID."""
        pass


class UserService(Protocol):
    """Protocol for user service operations."""

    def get_user(self, user_id: int) -> Optional[User]:
        """Get user by ID."""
        ...

    def create_user(self, name: str, email: str) -> User:
        """Create a new user."""
        ...

    def update_user(self, user: User) -> bool:
        """Update an existing user."""
        ...

    def delete_user(self, user_id: int) -> bool:
        """Delete user by ID."""
        ...


class InMemoryUserRepository(Repository):
    """In-memory implementation of Repository for Users."""

    def __init__(self):
        """Initialize the repository."""
        self._users: Dict[int, User] = {}
        self._next_id = 1

    def save(self, entity: User) -> bool:
        """Save a user entity."""
        if entity.id == 0:
            entity.id = self._next_id
            self._next_id += 1

        self._users[entity.id] = entity
        return True

    def find_by_id(self, entity_id: int) -> Optional[User]:
        """Find user by ID."""
        return self._users.get(entity_id)

    def find_all(self) -> List[User]:
        """Find all users."""
        return list(self._users.values())

    def delete(self, entity_id: int) -> bool:
        """Delete user by ID."""
        if entity_id in self._users:
            del self._users[entity_id]
            return True
        return False

    def get_user_count(self) -> int:
        """Get total number of users."""
        return len(self._users)


def create_default_config() -> Config:
    """Factory function to create default configuration."""
    return Config.create_default()
