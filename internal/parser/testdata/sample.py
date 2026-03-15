"""Sample Python module for parser testing."""

import os
from pathlib import Path
from typing import List, Optional


class UserModel:
    """Represents a user in the system."""

    def __init__(self, name: str, email: str) -> None:
        self.name = name
        self.email = email

    def display_name(self) -> str:
        """Return the user's display name."""
        return self.name

    def greet(self, greeting: str = "Hello") -> str:
        """Return a greeting for the user."""
        return f"{greeting}, {self.name}!"


class AdminUser(UserModel):
    """An admin user with elevated privileges."""

    def __init__(self, name: str, email: str, role: str = "admin") -> None:
        super().__init__(name, email)
        self.role = role


def process_request(
    path: str,
    method: str = "GET",
    headers: Optional[dict] = None,
) -> dict:
    """Process an incoming HTTP request.

    Args:
        path: The request path.
        method: HTTP method.
        headers: Optional headers dict.

    Returns:
        Response dictionary.
    """
    return {"path": path, "method": method}


def helper_func():
    pass


def test_process_request():
    """Test the process_request function."""
    result = process_request("/api/users")
    assert result["path"] == "/api/users"
