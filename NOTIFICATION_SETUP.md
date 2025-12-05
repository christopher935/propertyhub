# Real-Time Admin Notification System Setup

## Overview
This document describes the setup and integration of the real-time notification system for the PropertyHub admin dashboard.

## Components Implemented

### 1. Models
- **AdminNotification** (`internal/models/admin_notification.go`)
  - Stores notification records with type, priority, lead/property references
  - Helper methods for marking as read/dismissed
  - Icon and color getters for frontend display

### 2. Services
- **AdminNotificationService** (`internal/services/admin_notification_service.go`)
  - Core service for creating and managing notifications
  - Trigger methods for different notification types:
    - `OnHotLeadActive()` - Lead with score 70+ browsing
    - `OnApplicationSubmitted()` - New application
    - `OnBookingCreated()` - New tour booking
    - `OnReturnVisitor()` - Hot lead returns after 24+ hours
    - `OnEngagementSpike()` - Score jumps 20+ points
    - `OnPropertySaved()` - Lead saves property
    - `OnInquirySent()` - Contact form submission
    - `OnMultiplePropertiesViewed()` - 5+ properties in session

### 3. WebSocket Infrastructure
- **WebSocketHub** (`internal/handlers/websocket_handlers.go`)
  - Manages client connections
  - Broadcasts notifications to admin users
  - Handles ping/pong for connection health
  - Auto-reconnection on disconnect

### 4. API Handlers
- **AdminNotificationHandler** (`internal/handlers/admin_notification_handlers.go`)
  - `GET /api/admin/notifications` - List recent notifications
  - `GET /api/admin/notifications/unread-count` - Get unread count
  - `POST /api/admin/notifications/:id/read` - Mark as read
  - `POST /api/admin/notifications/:id/dismiss` - Dismiss notification
  - `POST /api/admin/notifications/dismiss-all` - Dismiss all

### 5. Frontend UI
- **Notification Bell** (`web/templates/shared/components/admin-header.html`)
  - Badge with unread count
  - Dropdown with recent notifications
  - WebSocket connection for real-time updates
  - Click notification to mark as read

- **Toast Notifications** (`web/templates/admin/pages/admin-dashboard.html`)
  - Slide-in toasts for high-priority notifications
  - Auto-dismiss after 10 seconds
  - Sound alert for high-priority (optional)
  - Click action button to navigate to relevant page

- **Styles** (`web/static/css/notifications.css`)
  - Notification bell and dropdown styles
  - Toast notification styles
  - Animations and transitions
  - Responsive design

## Integration Steps

### Step 1: Add AdminNotification to Database Migration

In your database initialization (likely in `cmd/server/main.go` or `internal/database/database.go`), add:

```go
import "chrisgross-ctrl-project/internal/models"

// After database connection is established
err := gormDB.AutoMigrate(
    &models.AdminNotification{},
    // ... other models
)
```

### Step 2: Initialize Services and Handlers

In `cmd/server/main.go`, after other service initializations:

```go
// Initialize WebSocket Hub
wsHub := handlers.NewWebSocketHub()
go wsHub.Run()
log.Println("🔌 WebSocket hub started")

// Initialize Admin Notification Service
adminNotificationService := services.NewAdminNotificationService(gormDB, wsHub)
log.Println("🔔 Admin notification service initialized")

// Initialize Admin Notification Handler
adminNotificationHandler := handlers.NewAdminNotificationHandler(gormDB, adminNotificationService)
log.Println("📬 Admin notification handler initialized")

// Wire up notification service to existing services
behavioralEventService.SetNotificationService(adminNotificationService)
bookingHandler.SetNotificationService(adminNotificationService)
applicationWorkflowHandler.SetNotificationService(adminNotificationService)
log.Println("✅ Notification triggers wired to existing services")
```

### Step 3: Add to AllHandlers Struct

In `cmd/server/handlers.go`, add the notification handler:

```go
type AllHandlers struct {
    // ... existing handlers
    
    // Notifications
    AdminNotification     *handlers.AdminNotificationHandler
    WebSocket             *handlers.WebSocketHandler
    
    // ... rest of handlers
}
```

Then in `cmd/server/main.go` where AllHandlers is initialized:

```go
allHandlers := &AllHandlers{
    // ... existing handlers
    AdminNotification: adminNotificationHandler,
    WebSocket:         handlers.NewWebSocketHandler(wsHub),
    // ... rest
}
```

### Step 4: Register Routes

In `cmd/server/routes_api.go`, add notification API routes:

```go
// Admin Notification API
api.GET("/admin/notifications", h.AdminNotification.GetNotifications)
api.GET("/admin/notifications/unread-count", h.AdminNotification.GetUnreadCount)
api.POST("/admin/notifications/:id/read", h.AdminNotification.MarkAsRead)
api.POST("/admin/notifications/:id/dismiss", h.AdminNotification.Dismiss)
api.POST("/admin/notifications/dismiss-all", h.AdminNotification.DismissAll)
```

In `cmd/server/routes_admin.go` (or wherever WebSocket routes go), add:

```go
// WebSocket endpoint for real-time notifications
r.GET("/ws/admin", h.WebSocket.HandleConnection)
```

### Step 5: Update Admin Templates

The frontend templates have already been updated:
- `web/templates/shared/components/admin-header.html` - Notification bell with dropdown
- `web/templates/admin/pages/admin-dashboard.html` - Toast notifications

Ensure the CSS is linked in all admin pages:

```html
<link rel="stylesheet" href="/static/css/notifications.css">
```

## Notification Triggers

### Automatic Triggers
The system automatically triggers notifications for:

1. **Hot Lead Active** - When a lead's score reaches 70+ (triggered in `BehavioralEventService.TrackEvent`)
2. **Engagement Spike** - When a lead's score increases by 20+ points in one session
3. **Multiple Properties Viewed** - When a lead views 5+ properties in one session
4. **Property Saved** - When a lead saves a property (triggered in `TrackPropertySave`)
5. **Inquiry Sent** - When a lead submits a contact form (triggered in `TrackInquiry`)
6. **Application Submitted** - When a new application is created (triggered in `CreateApplicationNumber`)
7. **Booking Created** - When a new tour booking is submitted (triggered in `CreateBooking`)

### Manual Triggers
You can manually trigger notifications from anywhere in the codebase:

```go
// Inject the notification service where needed
notificationService := services.NewAdminNotificationService(db, wsHub)

// Trigger a notification
err := notificationService.OnHotLeadActive(leadID, sessionID)
```

## Testing

### 1. Backend Testing
```bash
# Start the server
go run cmd/server/main.go

# Check WebSocket connection
# Open browser console on admin dashboard and verify WebSocket connection
```

### 2. Frontend Testing
1. Log into admin dashboard
2. Check notification bell in header
3. Open browser console and verify WebSocket connection: `WebSocket connected`
4. Trigger a notification by:
   - Creating a new booking
   - Submitting an application
   - Having a lead browse properties (if tracking is enabled)
5. Verify toast appears and notification count increases

### 3. Database Testing
```sql
-- Check notifications table
SELECT * FROM admin_notifications ORDER BY created_at DESC LIMIT 10;

-- Check unread notifications
SELECT COUNT(*) FROM admin_notifications WHERE read_at IS NULL AND dismissed_at IS NULL;
```

## Configuration

### Sound Alerts
Sound alerts are enabled by default for high-priority notifications. Users can disable them via browser localStorage:

```javascript
localStorage.setItem('notification_sound', 'false');
```

### Notification Retention
Notifications are soft-deleted after being dismissed. To clean up old notifications:

```sql
-- Delete notifications older than 30 days
DELETE FROM admin_notifications 
WHERE dismissed_at IS NOT NULL 
AND dismissed_at < NOW() - INTERVAL '30 days';
```

## Troubleshooting

### WebSocket Not Connecting
1. Check WebSocket endpoint is registered: `/ws/admin`
2. Verify WebSocket upgrade is allowed (check CORS/proxy settings)
3. Check browser console for connection errors

### Notifications Not Appearing
1. Verify database has `admin_notifications` table
2. Check notification service is wired to handlers
3. Check WebSocket hub is running: `go wsHub.Run()`
4. Verify user role is "admin" or "super_admin"

### Toast Not Showing
1. Ensure CSS is loaded: `/static/css/notifications.css`
2. Check Alpine.js is loaded
3. Verify `@notification-received` event listener is attached

## Future Enhancements

1. **Email Digest** - Daily email with notification summary
2. **Notification Preferences** - Per-user notification settings
3. **Notification History Page** - Full-page view of all notifications
4. **Notification Filters** - Filter by type, priority, date range
5. **Notification Templates** - Customizable notification messages
6. **Push Notifications** - Browser push notifications for high-priority alerts
7. **Mobile App Integration** - Push notifications to mobile app
8. **Slack Integration** - Send notifications to Slack channel

## Support
For issues or questions, contact the development team or create an issue in the repository.
