#!/usr/bin/env python3
"""Simple Python main module for testing AST parsing."""

import os
import sys
from typing import List, Dict, Optional, Union


def format_name(name: str) -> str:
    """Format a name by capitalizing first letter of each word."""
    return ' '.join(word.capitalize() for word in name.split())


def validate_email(email: str) -> bool:
    """Basic email validation."""
    return '@' in email and '.' in email.split('@')[1]


def process_user_data(name: str, email: str, age: int = 18) -> Dict[str, Union[str, int, bool]]:
    """Process user data and return formatted result."""
    if not name.strip():
        raise ValueError("Name cannot be empty")

    if not validate_email(email):
        raise ValueError("Invalid email format")

    formatted_name = format_name(name)

    return {
        'name': formatted_name,
        'email': email.lower(),
        'age': age,
        'is_adult': age >= 18,
        'status': 'active'
    }


def calculate_statistics(numbers: List[float]) -> Dict[str, float]:
    """Calculate basic statistics for a list of numbers."""
    if not numbers:
        return {'count': 0, 'sum': 0.0, 'mean': 0.0, 'min': 0.0, 'max': 0.0}

    count = len(numbers)
    total = sum(numbers)
    mean = total / count

    return {
        'count': count,
        'sum': total,
        'mean': mean,
        'min': min(numbers),
        'max': max(numbers)
    }


def search_users(users: List[Dict[str, str]], query: str) -> List[Dict[str, str]]:
    """Search users by name or email."""
    query_lower = query.lower()
    results = []

    for user in users:
        name = user.get('name', '').lower()
        email = user.get('email', '').lower()

        if query_lower in name or query_lower in email:
            results.append(user)

    return results


def generate_report(data: List[Dict[str, Union[str, int]]]) -> str:
    """Generate a simple text report from data."""
    if not data:
        return "No data available"

    report_lines = ["User Report", "=" * 20]

    for item in data:
        name = item.get('name', 'Unknown')
        age = item.get('age', 0)
        report_lines.append(f"Name: {name}, Age: {age}")

    return '\n'.join(report_lines)


def main():
    """Main function demonstrating the functionality."""
    print("=== Python Simple Demo ===")

    # Process some user data
    try:
        user_data = process_user_data("john doe", "john@example.com", 25)
        print(f"Processed user: {user_data}")
    except ValueError as e:
        print(f"Error processing user: {e}")

    # Calculate statistics
    numbers = [1.5, 2.8, 3.2, 4.1, 5.9, 2.3, 7.8]
    stats = calculate_statistics(numbers)
    print(f"Statistics: {stats}")

    # Search functionality
    users = [
        {'name': 'Alice Smith', 'email': 'alice@example.com'},
        {'name': 'Bob Johnson', 'email': 'bob@example.com'},
        {'name': 'Charlie Brown', 'email': 'charlie@example.com'}
    ]

    search_results = search_users(users, 'alice')
    print(f"Search results: {search_results}")

    # Generate report
    report_data = [
        {'name': 'Alice', 'age': 30},
        {'name': 'Bob', 'age': 25},
        {'name': 'Charlie', 'age': 35}
    ]

    report = generate_report(report_data)
    print(f"Report:\n{report}")


if __name__ == "__main__":
    main()
