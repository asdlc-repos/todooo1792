# Overview

A web-based task management application that allows authenticated users to create, organize, and track their personal to-do items. Users can categorize tasks, set due dates, and manage their task list through standard CRUD operations.

# Personas

- **Emma** — Individual User: A professional managing personal and work tasks who needs a simple, reliable way to track to-dos with deadlines and organization.
- **Admin** — System Administrator: Responsible for monitoring system health, managing user accounts, and ensuring data integrity.

# Capabilities

## User Authentication & Account Management

- The system SHALL require users to register with a unique email address and password.
- WHEN a user submits registration credentials, the system SHALL create an account and send a verification email.
- WHEN a user submits valid login credentials, the system SHALL authenticate the user and create a session.
- WHEN a user requests password reset, the system SHALL send a secure reset link to the registered email address.
- The system SHALL encrypt all passwords using industry-standard hashing algorithms.
- WHEN a user session expires after 24 hours of inactivity, the system SHALL log out the user automatically.
- The system SHALL allow users to update their email address and password through account settings.

## Task Management

- WHILE authenticated, the user SHALL be able to create a new task with a title and optional description.
- WHILE authenticated, the user SHALL be able to view all tasks they have created.
- WHILE authenticated, the user SHALL be able to edit the title, description, category, and due date of their existing tasks.
- WHILE authenticated, the user SHALL be able to delete tasks they have created.
- The system SHALL restrict each user to viewing and modifying only their own tasks.
- WHEN a user creates a task, the system SHALL assign a unique identifier and timestamp to that task.
- The system SHALL support task titles up to 200 characters and descriptions up to 2000 characters.

## Task Categorization

- The system SHALL allow users to assign one category to each task.
- WHILE authenticated, the user SHALL be able to create custom categories with unique names.
- WHILE authenticated, the user SHALL be able to edit and delete their custom categories.
- WHEN a user deletes a category, the system SHALL remove the category assignment from all associated tasks.
- The system SHALL allow users to filter their task list by one or more categories.
- The system SHALL support category names up to 50 characters.

## Due Dates & Scheduling

- The system SHALL allow users to set a due date and time for each task.
- The system SHALL allow users to create tasks without a due date.
- The system SHALL display tasks in chronological order by due date, with undated tasks appearing last.
- WHEN the current date matches a task's due date, the system SHALL visually highlight that task as due today.
- WHEN the current date exceeds a task's due date, the system SHALL visually highlight that task as overdue.
- The system SHALL allow users to filter tasks by date range (today, this week, this month, overdue).

## Task Status & Completion

- The system SHALL allow users to mark tasks as complete or incomplete.
- WHEN a user marks a task as complete, the system SHALL record the completion timestamp.
- The system SHALL allow users to filter tasks by completion status (all, active, completed).
- WHILE viewing completed tasks, the system SHALL display the completion timestamp for each task.

## User Interface & Interaction

- The system SHALL provide a responsive interface that functions on desktop and mobile browsers.
- The system SHALL display task counts by category and status on the main dashboard.
- WHEN a user performs a create, update, or delete operation, the system SHALL provide visual confirmation within 2 seconds.
- IF a network error occurs during a task operation, THEN the system SHALL display an error message and retain user input.
- The system SHALL allow users to search tasks by title and description using text matching.
- WHEN a user navigates between pages, the system SHALL preserve filter and sort selections.

## Data Persistence & Performance

- The system SHALL persist all task data, ensuring no data loss after user logout or session expiration.
- WHEN a user requests their task list, the system SHALL return results within 500 milliseconds for up to 1000 tasks.
- The system SHALL support concurrent access by multiple users without data corruption.
- The system SHALL back up all user data daily.

## Security & Privacy

- The system SHALL use HTTPS for all client-server communication.
- The system SHALL validate and sanitize all user inputs to prevent injection attacks.
- WHEN a user is not authenticated, the system SHALL redirect access attempts to protected pages to the login screen.
- The system SHALL implement rate limiting of 100 requests per minute per user to prevent abuse.
- IF authentication fails five times within 15 minutes, THEN the system SHALL temporarily lock the account for 30 minutes.