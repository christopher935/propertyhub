# Live Sessions Panel - "Who's Browsing Now" Implementation

## Overview
Created a real-time "Who's Browsing Now" panel on the admin dashboard that shows active consumer sessions with behavioral scores and activity.

## Components Implemented

### 1. Backend API Endpoints

**File**: `internal/handlers/behavioral_sessions_handlers.go`

#### GET `/api/admin/sessions/active`
Returns currently active sessions with enriched data:
- Session details (ID, start time, duration)
- Lead information (email, name, anonymous status)
- Location data (city, state)
- Behavioral score and category (hot/warm/cold)
- Current activity and property being viewed
- Session metrics (page views, property views, saves)
- Device and browser information

**Response Structure**:
```json
{
  "sessions": [...],
  "total_count": 5,
  "hot_count": 2,
  "warm_count": 2,
  "cold_count": 1
}
```

#### GET `/api/admin/sessions/:id/journey`
Returns complete session timeline:
- All events in chronological order
- Properties viewed with details
- Score changes during session
- Session summary statistics

### 2. Frontend Components

**File**: `web/templates/admin/pages/admin-dashboard.html`

#### Live Sessions Panel
- Displays up to 5 most recent active sessions
- Real-time updates via polling (10-second interval)
- Color-coded by behavioral score:
  - 🔴 Hot (70+): Red
  - 🟡 Warm (40-69): Orange
  - 🔵 Cold (<40): Blue

#### Session Card Features
- **Status Indicator**: Emoji badge showing temperature
- **Identity**: Email/name or "Anonymous Visitor"
- **Location**: City, State with location icon
- **Behavioral Score**: Numeric score with category label
- **Current Activity**: Page being viewed
- **Property Context**: Address of property being viewed
- **Session Metrics**:
  - Page views count
  - Session duration (formatted as minutes/seconds)
  - Device type
- **Quick Actions**:
  - View Journey: Opens timeline modal
  - Identify Lead: For anonymous visitors
  - Send Property: For identified leads

#### Session Journey Modal
- Full timeline of all session events
- Chronological activity list
- Property details for viewed properties
- Session summary statistics

### 3. Styling

**File**: `web/static/css/live-sessions.css`

Features:
- Pulse animation for live indicator
- Smooth transitions and hover effects
- Score-based color coding
- Responsive design (mobile-friendly)
- Loading and empty states
- Timeline visualization
- Modal overlay with backdrop

### 4. Real-Time Updates

Implementation:
- **Polling**: 10-second interval using `setInterval`
- **Update Animation**: New sessions fade in
- **Change Detection**: Highlights updated sessions
- **Auto-refresh**: Continues until component destroyed

### 5. Integration Points

**Files Modified**:
- `cmd/server/handlers.go` - Added BehavioralSessionsHandler to AllHandlers struct
- `cmd/server/main.go` - Initialized handler and added to routing
- `cmd/server/routes_api.go` - Registered API endpoints

## Data Sources

### Database Tables Used
1. **behavioral_sessions** - Session tracking data
2. **behavioral_scores** - Current behavioral scores
3. **behavioral_events** - Event history
4. **leads** - Lead information (email, name, location)
5. **properties** - Property details

### Session Criteria
Active sessions defined as:
- `end_time` IS NULL
- `start_time` >= 15 minutes ago
- Ordered by start_time DESC

## Behavioral Score Categories

| Score | Category | Color | Icon |
|-------|----------|-------|------|
| 70+   | Hot      | Red   | 🔴   |
| 40-69 | Warm     | Orange| 🟡   |
| 0-39  | Cold     | Blue  | 🔵   |

## Features Implemented

✅ Real-time session monitoring
✅ Behavioral score display with color coding
✅ Anonymous vs identified lead differentiation
✅ Current activity tracking
✅ Property being viewed display
✅ Session metrics (duration, page views, device)
✅ Location display (city, state)
✅ Session journey timeline modal
✅ Quick actions (view journey, identify, send property)
✅ Auto-refresh without page reload
✅ Responsive mobile design
✅ Loading and empty states
✅ Smooth animations and transitions

## API Response Example

```json
{
  "sessions": [
    {
      "session_id": "uuid-here",
      "lead_id": 123,
      "lead_email": "john@example.com",
      "lead_name": "John Doe",
      "is_anonymous": false,
      "location": {
        "city": "Houston",
        "state": "TX",
        "country": "USA"
      },
      "behavioral_score": 78,
      "score_category": "hot",
      "current_page": "Property Details",
      "current_property": {
        "id": 456,
        "address": "123 Main St",
        "city": "Houston",
        "state": "TX",
        "price": 250000
      },
      "page_views": 7,
      "property_views": 3,
      "property_saves": 1,
      "session_duration": 272,
      "last_activity": "2025-12-05T12:34:56Z",
      "start_time": "2025-12-05T12:30:00Z",
      "device_type": "desktop",
      "browser": "Chrome",
      "referrer_source": "google.com"
    }
  ],
  "total_count": 1,
  "hot_count": 1,
  "warm_count": 0,
  "cold_count": 0
}
```

## Future Enhancements

Potential improvements:
1. **WebSocket Integration** - Replace polling with WebSocket for true real-time updates
2. **Advanced Filtering** - Filter by score category, location, device type
3. **Session Replay** - Visual replay of user's browsing path
4. **Live Chat Integration** - Proactive chat with hot leads
5. **Property Recommendations** - AI-suggested properties based on browsing
6. **Lead Scoring Trends** - Show score changes over session duration
7. **Geographic Clustering** - Map view of active sessions
8. **Notification System** - Alert when high-value leads become active

## Testing Checklist

- [ ] Backend API returns active sessions correctly
- [ ] Session enrichment includes all required data
- [ ] Behavioral scores calculated accurately
- [ ] Location data displayed properly
- [ ] Anonymous vs identified leads differentiated
- [ ] Current property context shown correctly
- [ ] Session journey modal displays timeline
- [ ] Polling updates every 10 seconds
- [ ] Color coding matches score categories
- [ ] Actions (view journey, identify, send) work
- [ ] Responsive design works on mobile
- [ ] Empty state displayed when no sessions
- [ ] Loading state shows during fetch

## Design System Compliance

Uses existing design tokens:
- `--navy-primary` - Headers and primary text
- `--gold-primary` - Accents and highlights
- `--hot` (#EF4444) - Hot lead color
- `--warm` (#F59E0B) - Warm lead color
- `--cold` (#3B82F6) - Cold lead color
- `--font-family-primary` - Playfair Display for titles
- `--font-family-body` - Source Sans 3 for body text
- `--font-family-mono` - IBM Plex Mono for metrics

## Performance Considerations

- **Polling Interval**: 10 seconds (configurable)
- **Session Limit**: Top 5 displayed on dashboard
- **Database Query**: Optimized with proper indexes
- **Caching**: Consider Redis caching for high traffic
- **Lazy Loading**: Journey modal data fetched on demand

## Deployment Notes

1. Ensure behavioral_sessions table exists and is populated
2. Verify behavioral_scores are being calculated
3. Check that session tracking is enabled on consumer site
4. Test with both anonymous and authenticated users
5. Monitor performance with multiple concurrent sessions
6. Configure polling interval based on server capacity

## File Changes Summary

**New Files**:
- `internal/handlers/behavioral_sessions_handlers.go` (395 lines)
- `web/static/css/live-sessions.css` (621 lines)

**Modified Files**:
- `cmd/server/handlers.go` (+1 line)
- `cmd/server/main.go` (+2 lines)
- `cmd/server/routes_api.go` (+3 lines)
- `web/templates/admin/pages/admin-dashboard.html` (+174 lines)

**Total**: 1,196 lines added

## Git Commit

Branch: `capy/build-live-session-d-2e8efc42`
Commit: `e45af8c`
Message: "feat: Add Who's Browsing Now live sessions panel to admin dashboard"
