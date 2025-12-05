/**
 * PropertyHub Calendar System
 * Availability calendar, booking scheduler, blackout dates management
 * Property inspection scheduling with TREC compliance
 */

class PropertyHubCalendar {
    constructor() {
        this.calendars = new Map();
        this.events = [];
        this.blackoutDates = [];
        this.availabilityData = {};
        this.selectedDates = {};
        this.bookingRules = {};
        this.currentView = 'month';
        this.currentDate = new Date();
        this.timeSlots = [];
        
        // Calendar types
        this.calendarTypes = {
            AVAILABILITY: 'availability',
            BOOKING: 'booking',
            INSPECTION: 'inspection',
            MAINTENANCE: 'maintenance',
            SHOWINGS: 'showings',
            CLOSINGS: 'closings'
        };

        // Event types
        this.eventTypes = {
            BOOKING: 'booking',
            INSPECTION: 'inspection',
            SHOWING: 'showing',
            MAINTENANCE: 'maintenance',
            CLOSING: 'closing',
            BLACKOUT: 'blackout',
            AVAILABLE: 'available',
            UNAVAILABLE: 'unavailable'
        };

        // Time slot configuration
        this.timeSlotConfig = {
            startHour: 8,
            endHour: 20,
            duration: 60, // minutes
            bufferTime: 30, // minutes between appointments
            workDays: [1, 2, 3, 4, 5, 6], // Monday to Saturday
            timezone: 'America/Chicago'
        };

        this.init();
    }

    async init() {
        console.log('📅 PropertyHub Calendar initializing...');
        
        await this.loadCalendarData();
        this.setupCalendarInstances();
        this.generateTimeSlots();
        this.setupEventListeners();
        
        console.log('✅ PropertyHub Calendar ready');
    }

    // Calendar Instance Management
    setupCalendarInstances() {
        const calendarElements = document.querySelectorAll('.property-calendar');
        
        calendarElements.forEach(element => {
            this.initializeCalendar(element);
        });
    }

    initializeCalendar(element) {
        const calendarId = element.id || this.generateCalendarId();
        const calendarType = element.dataset.type || this.calendarTypes.AVAILABILITY;
        const propertyId = element.dataset.propertyId;
        const allowMultiple = element.dataset.allowMultiple === 'true';
        const minDate = element.dataset.minDate ? new Date(element.dataset.minDate) : new Date();
        const maxDate = element.dataset.maxDate ? new Date(element.dataset.maxDate) : null;

        const calendarConfig = {
            element: element,
            id: calendarId,
            type: calendarType,
            propertyId: propertyId,
            allowMultiple: allowMultiple,
            minDate: minDate,
            maxDate: maxDate,
            selectedDates: [],
            events: []
        };

        this.calendars.set(calendarId, calendarConfig);
        this.renderCalendar(calendarId);
    }

    renderCalendar(calendarId) {
        const config = this.calendars.get(calendarId);
        if (!config) return;

        const element = config.element;
        element.innerHTML = `
            <div class="calendar-container">
                <div class="calendar-header">
                    <div class="calendar-navigation">
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubCalendar.navigateMonth('${calendarId}', -1)">
                            <i class="fas fa-chevron-left"></i>
                        </button>
                        <h3 class="calendar-title" id="calendar-title-${calendarId}">
                            ${this.formatMonthYear(this.currentDate)}
                        </h3>
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubCalendar.navigateMonth('${calendarId}', 1)">
                            <i class="fas fa-chevron-right"></i>
                        </button>
                    </div>
                    
                    <div class="calendar-views">
                        <button class="view-btn ${this.currentView === 'month' ? 'active' : ''}" 
                                onclick="PropertyHubCalendar.changeView('${calendarId}', 'month')">
                            Month
                        </button>
                        <button class="view-btn ${this.currentView === 'week' ? 'active' : ''}" 
                                onclick="PropertyHubCalendar.changeView('${calendarId}', 'week')">
                            Week
                        </button>
                        ${config.type === this.calendarTypes.INSPECTION ? `
                            <button class="view-btn ${this.currentView === 'day' ? 'active' : ''}" 
                                    onclick="PropertyHubCalendar.changeView('${calendarId}', 'day')">
                                Day
                            </button>
                        ` : ''}
                    </div>
                </div>

                <div class="calendar-legend" id="calendar-legend-${calendarId}">
                    ${this.renderCalendarLegend(config.type)}
                </div>

                <div class="calendar-grid" id="calendar-grid-${calendarId}">
                    ${this.renderCalendarGrid(calendarId)}
                </div>

                ${config.type === this.calendarTypes.INSPECTION ? `
                    <div class="time-slots-container" id="time-slots-${calendarId}" style="display: none;">
                        <h4>Available Time Slots</h4>
                        <div class="time-slots-grid" id="time-slots-grid-${calendarId}">
                            <!-- Time slots will be populated when a date is selected -->
                        </div>
                    </div>
                ` : ''}

                <div class="calendar-actions">
                    ${this.renderCalendarActions(config)}
                </div>
            </div>
        `;
    }

    renderCalendarGrid(calendarId) {
        if (this.currentView === 'month') {
            return this.renderMonthView(calendarId);
        } else if (this.currentView === 'week') {
            return this.renderWeekView(calendarId);
        } else if (this.currentView === 'day') {
            return this.renderDayView(calendarId);
        }
    }

    renderMonthView(calendarId) {
        const config = this.calendars.get(calendarId);
        const year = this.currentDate.getFullYear();
        const month = this.currentDate.getMonth();
        
        // Get first day of month and calculate calendar grid
        const firstDay = new Date(year, month, 1);
        const lastDay = new Date(year, month + 1, 0);
        const startDate = new Date(firstDay);
        startDate.setDate(startDate.getDate() - firstDay.getDay()); // Start from Sunday
        
        const weeks = [];
        let currentWeek = [];
        
        for (let i = 0; i < 42; i++) { // 6 weeks * 7 days
            const date = new Date(startDate);
            date.setDate(startDate.getDate() + i);
            
            currentWeek.push(date);
            
            if (currentWeek.length === 7) {
                weeks.push(currentWeek);
                currentWeek = [];
            }
        }

        return `
            <div class="calendar-month-view">
                <div class="calendar-weekdays">
                    <div class="weekday">Sun</div>
                    <div class="weekday">Mon</div>
                    <div class="weekday">Tue</div>
                    <div class="weekday">Wed</div>
                    <div class="weekday">Thu</div>
                    <div class="weekday">Fri</div>
                    <div class="weekday">Sat</div>
                </div>
                
                <div class="calendar-weeks">
                    ${weeks.map(week => `
                        <div class="calendar-week">
                            ${week.map(date => this.renderCalendarDay(calendarId, date, month)).join('')}
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    renderCalendarDay(calendarId, date, currentMonth) {
        const config = this.calendars.get(calendarId);
        const dateStr = this.formatDateString(date);
        const isCurrentMonth = date.getMonth() === currentMonth;
        const isToday = this.isToday(date);
        const isPast = this.isPastDate(date);
        const isSelected = config.selectedDates.includes(dateStr);
        const isBlackout = this.isBlackoutDate(config.propertyId, date);
        const availability = this.getDateAvailability(config.propertyId, date);
        const events = this.getDateEvents(config.propertyId, date);

        let dayClasses = ['calendar-day'];
        
        if (!isCurrentMonth) dayClasses.push('other-month');
        if (isToday) dayClasses.push('today');
        if (isPast) dayClasses.push('past');
        if (isSelected) dayClasses.push('selected');
        if (isBlackout) dayClasses.push('blackout');
        
        // Add availability classes
        if (availability) {
            dayClasses.push(`availability-${availability.status}`);
        }

        // Check if date is selectable
        const isSelectable = this.isDateSelectable(config, date);
        if (isSelectable) {
            dayClasses.push('selectable');
        }

        return `
            <div class="${dayClasses.join(' ')}" 
                 data-date="${dateStr}"
                 data-calendar="${calendarId}"
                 ${isSelectable ? `onclick="PropertyHubCalendar.selectDate('${calendarId}', '${dateStr}')"` : ''}>
                
                <div class="day-number">${date.getDate()}</div>
                
                ${events.length > 0 ? `
                    <div class="day-events">
                        ${events.slice(0, 3).map(event => `
                            <div class="event-indicator ${event.type}" title="${event.title}">
                                ${event.title.substring(0, 10)}...
                            </div>
                        `).join('')}
                        ${events.length > 3 ? `<div class="more-events">+${events.length - 3} more</div>` : ''}
                    </div>
                ` : ''}
                
                ${availability && availability.price ? `
                    <div class="day-price">$${availability.price}</div>
                ` : ''}
                
                ${isBlackout ? `<div class="blackout-indicator"><i class="fas fa-times"></i></div>` : ''}
            </div>
        `;
    }

    renderWeekView(calendarId) {
        const config = this.calendars.get(calendarId);
        const startOfWeek = this.getStartOfWeek(this.currentDate);
        const days = [];
        
        for (let i = 0; i < 7; i++) {
            const date = new Date(startOfWeek);
            date.setDate(startOfWeek.getDate() + i);
            days.push(date);
        }

        return `
            <div class="calendar-week-view">
                <div class="week-header">
                    ${days.map(date => `
                        <div class="week-day-header">
                            <div class="day-name">${this.formatDayName(date)}</div>
                            <div class="day-date">${date.getDate()}</div>
                        </div>
                    `).join('')}
                </div>
                
                <div class="week-body">
                    ${days.map(date => `
                        <div class="week-day-column" data-date="${this.formatDateString(date)}">
                            ${this.renderDaySchedule(calendarId, date)}
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    renderDayView(calendarId) {
        const config = this.calendars.get(calendarId);
        
        return `
            <div class="calendar-day-view">
                <div class="day-schedule">
                    ${this.renderDaySchedule(calendarId, this.currentDate)}
                </div>
            </div>
        `;
    }

    renderDaySchedule(calendarId, date) {
        const config = this.calendars.get(calendarId);
        const timeSlots = this.generateDayTimeSlots(date);
        const events = this.getDateEvents(config.propertyId, date);

        return `
            <div class="day-time-slots">
                ${timeSlots.map(slot => {
                    const slotEvents = events.filter(event => 
                        this.isTimeSlotOverlapping(slot, event)
                    );
                    
                    return `
                        <div class="time-slot ${slotEvents.length > 0 ? 'has-events' : 'available'}"
                             data-time="${slot.start}"
                             data-date="${this.formatDateString(date)}"
                             data-calendar="${calendarId}"
                             onclick="PropertyHubCalendar.selectTimeSlot('${calendarId}', '${this.formatDateString(date)}', '${slot.start}')">
                            
                            <div class="slot-time">${this.formatTime(slot.start)}</div>
                            
                            ${slotEvents.map(event => `
                                <div class="slot-event ${event.type}">
                                    <div class="event-title">${event.title}</div>
                                    <div class="event-details">${event.details || ''}</div>
                                </div>
                            `).join('')}
                        </div>
                    `;
                }).join('')}
            </div>
        `;
    }

    renderCalendarLegend(calendarType) {
        const legendItems = {
            [this.calendarTypes.AVAILABILITY]: [
                { class: 'available', label: 'Available', color: '#10b981' },
                { class: 'unavailable', label: 'Unavailable', color: '#ef4444' },
                { class: 'blackout', label: 'Blackout Date', color: '#6b7280' },
                { class: 'booked', label: 'Booked', color: '#f59e0b' }
            ],
            [this.calendarTypes.INSPECTION]: [
                { class: 'available', label: 'Available', color: '#10b981' },
                { class: 'scheduled', label: 'Scheduled', color: '#3b82f6' },
                { class: 'unavailable', label: 'Unavailable', color: '#ef4444' }
            ],
            [this.calendarTypes.BOOKING]: [
                { class: 'available', label: 'Available', color: '#10b981' },
                { class: 'booked', label: 'Booked', color: '#f59e0b' },
                { class: 'pending', label: 'Pending', color: '#8b5cf6' },
                { class: 'maintenance', label: 'Maintenance', color: '#ef4444' }
            ]
        };

        const items = legendItems[calendarType] || legendItems[this.calendarTypes.AVAILABILITY];
        
        return `
            <div class="legend-items">
                ${items.map(item => `
                    <div class="legend-item">
                        <div class="legend-color" style="background-color: ${item.color}"></div>
                        <span class="legend-label">${item.label}</span>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderCalendarActions(config) {
        switch (config.type) {
            case this.calendarTypes.AVAILABILITY:
                return `
                    <button class="btn btn-primary" onclick="PropertyHubCalendar.setAvailability('${config.id}')">
                        Set Availability
                    </button>
                    <button class="btn btn-secondary" onclick="PropertyHubCalendar.addBlackoutDate('${config.id}')">
                        Add Blackout Date
                    </button>
                `;
            
            case this.calendarTypes.INSPECTION:
                return `
                    <button class="btn btn-primary" id="schedule-inspection-${config.id}" 
                            onclick="PropertyHubCalendar.scheduleInspection('${config.id}')" 
                            disabled>
                        Schedule Inspection
                    </button>
                `;
            
            case this.calendarTypes.BOOKING:
                return `
                    <button class="btn btn-primary" id="confirm-booking-${config.id}" 
                            onclick="PropertyHubCalendar.confirmBooking('${config.id}')" 
                            disabled>
                        Confirm Booking
                    </button>
                `;
            
            default:
                return '';
        }
    }

    // Time Slot Management
    generateTimeSlots() {
        const { startHour, endHour, duration } = this.timeSlotConfig;
        const slots = [];

        for (let hour = startHour; hour < endHour; hour++) {
            for (let minute = 0; minute < 60; minute += duration) {
                const startTime = `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`;
                const endMinute = minute + duration;
                const endHour = hour + Math.floor(endMinute / 60);
                const endMin = endMinute % 60;
                const endTime = `${endHour.toString().padStart(2, '0')}:${endMin.toString().padStart(2, '0')}`;
                
                slots.push({
                    start: startTime,
                    end: endTime,
                    duration: duration
                });
            }
        }

        this.timeSlots = slots;
    }

    generateDayTimeSlots(date) {
        // Filter time slots based on availability and existing bookings
        const dateStr = this.formatDateString(date);
        const dayOfWeek = date.getDay();
        
        // Check if it's a work day
        if (!this.timeSlotConfig.workDays.includes(dayOfWeek)) {
            return [];
        }

        return this.timeSlots;
    }

    showTimeSlots(calendarId, date) {
        const config = this.calendars.get(calendarId);
        const timeSlotsContainer = document.getElementById(`time-slots-${calendarId}`);
        const timeSlotsGrid = document.getElementById(`time-slots-grid-${calendarId}`);
        
        if (!timeSlotsContainer || !timeSlotsGrid) return;

        const timeSlots = this.generateDayTimeSlots(date);
        const events = this.getDateEvents(config.propertyId, date);

        timeSlotsGrid.innerHTML = timeSlots.map(slot => {
            const isAvailable = !events.some(event => 
                this.isTimeSlotOverlapping(slot, event)
            );
            
            return `
                <button class="time-slot-btn ${isAvailable ? 'available' : 'unavailable'}"
                        data-time="${slot.start}"
                        ${isAvailable ? `onclick="PropertyHubCalendar.selectTimeSlot('${calendarId}', '${this.formatDateString(date)}', '${slot.start}')"` : ''}
                        ${!isAvailable ? 'disabled' : ''}>
                    ${this.formatTime(slot.start)} - ${this.formatTime(slot.end)}
                </button>
            `;
        }).join('');

        timeSlotsContainer.style.display = 'block';
    }

    // Date Selection and Availability
    selectDate(calendarId, dateStr) {
        const config = this.calendars.get(calendarId);
        if (!config) return;

        const date = new Date(dateStr);
        
        // Check if date is selectable
        if (!this.isDateSelectable(config, date)) {
            return;
        }

        if (config.allowMultiple) {
            // Toggle selection for multiple dates
            const index = config.selectedDates.indexOf(dateStr);
            if (index > -1) {
                config.selectedDates.splice(index, 1);
            } else {
                config.selectedDates.push(dateStr);
            }
        } else {
            // Single date selection
            config.selectedDates = [dateStr];
        }

        // Update UI
        this.updateCalendarSelection(calendarId);
        
        // Show time slots for inspection calendar
        if (config.type === this.calendarTypes.INSPECTION && config.selectedDates.length > 0) {
            this.showTimeSlots(calendarId, date);
            this.updateScheduleButton(calendarId);
        }

        // Update booking button
        if (config.type === this.calendarTypes.BOOKING) {
            this.updateBookingButton(calendarId);
        }
    }

    selectTimeSlot(calendarId, dateStr, timeStr) {
        const config = this.calendars.get(calendarId);
        if (!config) return;

        config.selectedTime = timeStr;
        
        // Update time slot selection UI
        const timeSlotsGrid = document.getElementById(`time-slots-grid-${calendarId}`);
        if (timeSlotsGrid) {
            timeSlotsGrid.querySelectorAll('.time-slot-btn').forEach(btn => {
                btn.classList.remove('selected');
            });
            
            const selectedBtn = timeSlotsGrid.querySelector(`[data-time="${timeStr}"]`);
            if (selectedBtn) {
                selectedBtn.classList.add('selected');
            }
        }

        // Update schedule button
        this.updateScheduleButton(calendarId);
    }

    isDateSelectable(config, date) {
        // Check minimum date
        if (config.minDate && date < config.minDate) {
            return false;
        }

        // Check maximum date
        if (config.maxDate && date > config.maxDate) {
            return false;
        }

        // Check if past date
        if (this.isPastDate(date)) {
            return false;
        }

        // Check blackout dates
        if (this.isBlackoutDate(config.propertyId, date)) {
            return false;
        }

        // Check availability
        const availability = this.getDateAvailability(config.propertyId, date);
        if (availability && availability.status === 'unavailable') {
            return false;
        }

        return true;
    }

    updateCalendarSelection(calendarId) {
        const config = this.calendars.get(calendarId);
        const calendarGrid = document.getElementById(`calendar-grid-${calendarId}`);
        
        if (!calendarGrid) return;

        // Remove all selected classes
        calendarGrid.querySelectorAll('.calendar-day').forEach(day => {
            day.classList.remove('selected');
        });

        // Add selected class to selected dates
        config.selectedDates.forEach(dateStr => {
            const dayElement = calendarGrid.querySelector(`[data-date="${dateStr}"]`);
            if (dayElement) {
                dayElement.classList.add('selected');
            }
        });
    }

    // Availability Management
    async setAvailability(calendarId) {
        const config = this.calendars.get(calendarId);
        if (!config || config.selectedDates.length === 0) {
            alert('Please select dates to set availability');
            return;
        }

        // Show availability modal
        this.showAvailabilityModal(calendarId, config.selectedDates);
    }

    showAvailabilityModal(calendarId, selectedDates) {
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal availability-modal">
                <div class="modal-header">
                    <h3>Set Availability</h3>
                    <button class="modal-close" onclick="this.closest('.modal-overlay').remove(); document.body.style.overflow = '';">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                
                <div class="modal-content">
                    <div class="selected-dates">
                        <h4>Selected Dates:</h4>
                        <div class="date-list">
                            ${selectedDates.map(date => `
                                <span class="date-tag">${new Date(date).toLocaleDateString()}</span>
                            `).join('')}
                        </div>
                    </div>
                    
                    <form id="availability-form">
                        <div class="form-group">
                            <label>Status</label>
                            <select name="status" class="form-control" required>
                                <option value="available">Available</option>
                                <option value="unavailable">Unavailable</option>
                                <option value="maintenance">Maintenance</option>
                            </select>
                        </div>
                        
                        <div class="form-group">
                            <label>Price per Night (optional)</label>
                            <input type="number" name="price" class="form-control" min="0" step="0.01">
                        </div>
                        
                        <div class="form-group">
                            <label>Minimum Stay (nights)</label>
                            <input type="number" name="minStay" class="form-control" min="1" value="1">
                        </div>
                        
                        <div class="form-group">
                            <label>Notes (optional)</label>
                            <textarea name="notes" class="form-control" rows="3"></textarea>
                        </div>
                    </form>
                </div>
                
                <div class="modal-actions">
                    <button class="btn btn-primary" onclick="PropertyHubCalendar.saveAvailability('${calendarId}', this)">
                        Save Availability
                    </button>
                    <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove(); document.body.style.overflow = '';">
                        Cancel
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(modal);
        document.body.style.overflow = 'hidden';
        
        // ESC key to close
        const escHandler = (e) => {
            if (e.key === 'Escape' && modal.parentElement) {
                modal.remove();
                document.body.style.overflow = '';
                document.removeEventListener('keydown', escHandler);
            }
        };
        document.addEventListener('keydown', escHandler);
        
        // Backdrop click to close
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
                document.body.style.overflow = '';
                document.removeEventListener('keydown', escHandler);
            }
        });
    }

    async saveAvailability(calendarId, buttonElement) {
        const config = this.calendars.get(calendarId);
        const modal = buttonElement.closest('.modal');
        const form = modal.querySelector('#availability-form');
        const formData = new FormData(form);

        const availabilityData = {
            propertyId: config.propertyId,
            dates: config.selectedDates,
            status: formData.get('status'),
            price: formData.get('price') ? parseFloat(formData.get('price')) : null,
            minStay: parseInt(formData.get('minStay')),
            notes: formData.get('notes')
        };

        try {
            buttonElement.disabled = true;
            buttonElement.textContent = 'Saving...';

            const response = await fetch('/api/properties/availability', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify(availabilityData)
            });

            if (!response.ok) {
                throw new Error('Failed to save availability');
            }

            // Update local data
            config.selectedDates.forEach(dateStr => {
                this.availabilityData[`${config.propertyId}_${dateStr}`] = availabilityData;
            });

            // Refresh calendar
            this.renderCalendar(calendarId);
            
            // Close modal
            modal.closest('.modal-overlay').remove();
            
            // Show success message
            this.showSuccessMessage('Availability updated successfully');

        } catch (error) {
            console.error('Error saving availability:', error);
            alert('Failed to save availability. Please try again.');
        } finally {
            buttonElement.disabled = false;
            buttonElement.textContent = 'Save Availability';
        }
    }

    // Blackout Date Management
    async addBlackoutDate(calendarId) {
        const config = this.calendars.get(calendarId);
        if (!config || config.selectedDates.length === 0) {
            alert('Please select dates to blackout');
            return;
        }

        const reason = prompt('Reason for blackout (optional):');
        
        try {
            const response = await fetch('/api/properties/blackout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    propertyId: config.propertyId,
                    dates: config.selectedDates,
                    reason: reason
                })
            });

            if (!response.ok) {
                throw new Error('Failed to add blackout dates');
            }

            // Update local blackout data
            config.selectedDates.forEach(dateStr => {
                this.blackoutDates.push({
                    propertyId: config.propertyId,
                    date: dateStr,
                    reason: reason
                });
            });

            // Refresh calendar
            this.renderCalendar(calendarId);
            
            this.showSuccessMessage('Blackout dates added successfully');

        } catch (error) {
            console.error('Error adding blackout dates:', error);
            alert('Failed to add blackout dates. Please try again.');
        }
    }

    // Inspection Scheduling
    async scheduleInspection(calendarId) {
        const config = this.calendars.get(calendarId);
        if (!config || config.selectedDates.length === 0 || !config.selectedTime) {
            alert('Please select a date and time for the inspection');
            return;
        }

        const inspectionData = {
            propertyId: config.propertyId,
            date: config.selectedDates[0],
            time: config.selectedTime,
            type: 'general_inspection',
            inspector: 'TBD',
            duration: 60 // minutes
        };

        try {
            const response = await fetch('/api/inspections/schedule', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify(inspectionData)
            });

            if (!response.ok) {
                throw new Error('Failed to schedule inspection');
            }

            const result = await response.json();
            
            // Add to events
            this.events.push({
                id: result.id,
                propertyId: config.propertyId,
                date: inspectionData.date,
                time: inspectionData.time,
                type: this.eventTypes.INSPECTION,
                title: 'Property Inspection',
                details: `Inspector: ${inspectionData.inspector}`
            });

            // Refresh calendar
            this.renderCalendar(calendarId);
            
            // Show confirmation
            this.showInspectionConfirmation(result);

        } catch (error) {
            console.error('Error scheduling inspection:', error);
            alert('Failed to schedule inspection. Please try again.');
        }
    }

    showInspectionConfirmation(inspection) {
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal inspection-confirmation">
                <div class="modal-header">
                    <h3>Inspection Scheduled</h3>
                    <button class="modal-close" onclick="this.closest('.modal-overlay').remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                
                <div class="modal-content">
                    <div class="confirmation-icon">
                        <i class="fas fa-check-circle fa-3x text-success"></i>
                    </div>
                    
                    <div class="inspection-details">
                        <h4>Your inspection has been scheduled!</h4>
                        <div class="detail-row">
                            <strong>Date:</strong> ${new Date(inspection.date).toLocaleDateString()}
                        </div>
                        <div class="detail-row">
                            <strong>Time:</strong> ${this.formatTime(inspection.time)}
                        </div>
                        <div class="detail-row">
                            <strong>Inspector:</strong> ${inspection.inspector || 'TBD'}
                        </div>
                        <div class="detail-row">
                            <strong>Confirmation #:</strong> ${inspection.id}
                        </div>
                    </div>
                    
                    <div class="next-steps">
                        <h5>What's Next:</h5>
                        <ul>
                            <li>You'll receive a confirmation email shortly</li>
                            <li>The inspector will contact you 24 hours before the appointment</li>
                            <li>Please ensure the property is accessible</li>
                            <li>Inspection typically takes 1-2 hours</li>
                        </ul>
                    </div>
                </div>
                
                <div class="modal-actions">
                    <button class="btn btn-primary" onclick="this.closest('.modal-overlay').remove()">
                        Got it, thanks!
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(modal);
    }

    // Navigation and View Management
    navigateMonth(calendarId, direction) {
        this.currentDate.setMonth(this.currentDate.getMonth() + direction);
        this.renderCalendar(calendarId);
    }

    changeView(calendarId, view) {
        this.currentView = view;
        this.renderCalendar(calendarId);
    }

    updateScheduleButton(calendarId) {
        const config = this.calendars.get(calendarId);
        const button = document.getElementById(`schedule-inspection-${calendarId}`);
        
        if (button) {
            button.disabled = !(config.selectedDates.length > 0 && config.selectedTime);
        }
    }

    updateBookingButton(calendarId) {
        const config = this.calendars.get(calendarId);
        const button = document.getElementById(`confirm-booking-${calendarId}`);
        
        if (button) {
            button.disabled = config.selectedDates.length === 0;
        }
    }

    // Data Loading and Management
    async loadCalendarData() {
        // Load availability data, events, blackout dates, etc.
        try {
            const [availabilityResponse, eventsResponse, blackoutResponse] = await Promise.all([
                fetch('/api/properties/availability', {
                    headers: { 'Authorization': `Bearer ${window.PropertyHubAuth?.token}` }
                }),
                fetch('/api/calendar/events', {
                    headers: { 'Authorization': `Bearer ${window.PropertyHubAuth?.token}` }
                }),
                fetch('/api/properties/blackout', {
                    headers: { 'Authorization': `Bearer ${window.PropertyHubAuth?.token}` }
                })
            ]);

            if (availabilityResponse.ok) {
                const availabilityData = await availabilityResponse.json();
                this.availabilityData = availabilityData;
            }

            if (eventsResponse.ok) {
                const eventsData = await eventsResponse.json();
                this.events = eventsData.events || [];
            }

            if (blackoutResponse.ok) {
                const blackoutData = await blackoutResponse.json();
                this.blackoutDates = blackoutData.blackoutDates || [];
            }

        } catch (error) {
            console.error('Error loading calendar data:', error);
        }
    }

    // Utility Methods
    generateCalendarId() {
        return 'calendar-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    formatDateString(date) {
        return date.toISOString().split('T')[0];
    }

    formatMonthYear(date) {
        return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
    }

    formatDayName(date) {
        return date.toLocaleDateString('en-US', { weekday: 'short' });
    }

    formatTime(timeStr) {
        const [hours, minutes] = timeStr.split(':');
        const hour = parseInt(hours);
        const ampm = hour >= 12 ? 'PM' : 'AM';
        const displayHour = hour > 12 ? hour - 12 : (hour === 0 ? 12 : hour);
        return `${displayHour}:${minutes} ${ampm}`;
    }

    isToday(date) {
        const today = new Date();
        return date.toDateString() === today.toDateString();
    }

    isPastDate(date) {
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        date = new Date(date);
        date.setHours(0, 0, 0, 0);
        return date < today;
    }

    getStartOfWeek(date) {
        const startOfWeek = new Date(date);
        startOfWeek.setDate(date.getDate() - date.getDay());
        return startOfWeek;
    }

    isBlackoutDate(propertyId, date) {
        const dateStr = this.formatDateString(date);
        return this.blackoutDates.some(blackout => 
            blackout.propertyId === propertyId && blackout.date === dateStr
        );
    }

    getDateAvailability(propertyId, date) {
        const dateStr = this.formatDateString(date);
        return this.availabilityData[`${propertyId}_${dateStr}`];
    }

    getDateEvents(propertyId, date) {
        const dateStr = this.formatDateString(date);
        return this.events.filter(event => 
            event.propertyId === propertyId && event.date === dateStr
        );
    }

    isTimeSlotOverlapping(slot, event) {
        if (!event.time) return false;
        
        const slotStart = this.timeToMinutes(slot.start);
        const slotEnd = this.timeToMinutes(slot.end);
        const eventStart = this.timeToMinutes(event.time);
        const eventEnd = eventStart + (event.duration || 60);
        
        return (slotStart < eventEnd && slotEnd > eventStart);
    }

    timeToMinutes(timeStr) {
        const [hours, minutes] = timeStr.split(':').map(Number);
        return hours * 60 + minutes;
    }

    showSuccessMessage(message) {
        if (window.PropertyHubNotifications) {
            window.PropertyHubNotifications.showToastNotification({
                title: 'Success',
                message: message,
                type: 'success',
                duration: 3000
            });
        }
    }

    setupEventListeners() {
        // Handle calendar events
        document.addEventListener('click', (e) => {
            if (e.target.matches('.calendar-day.selectable')) {
                const calendarId = e.target.dataset.calendar;
                const dateStr = e.target.dataset.date;
                this.selectDate(calendarId, dateStr);
            }
        });
    }

    // Public API Methods - These are called from the HTML onclick handlers
    static navigateMonth(calendarId, direction) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.navigateMonth(calendarId, direction);
        }
    }

    static changeView(calendarId, view) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.changeView(calendarId, view);
        }
    }

    static selectDate(calendarId, dateStr) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.selectDate(calendarId, dateStr);
        }
    }

    static selectTimeSlot(calendarId, dateStr, timeStr) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.selectTimeSlot(calendarId, dateStr, timeStr);
        }
    }

    static setAvailability(calendarId) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.setAvailability(calendarId);
        }
    }

    static addBlackoutDate(calendarId) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.addBlackoutDate(calendarId);
        }
    }

    static scheduleInspection(calendarId) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.scheduleInspection(calendarId);
        }
    }

    static confirmBooking(calendarId) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.confirmBooking(calendarId);
        }
    }

    static saveAvailability(calendarId, buttonElement) {
        if (window.PropertyHubCalendar) {
            window.PropertyHubCalendar.saveAvailability(calendarId, buttonElement);
        }
    }
}

// Initialize PropertyHub Calendar when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    if (document.querySelector('.property-calendar')) {
        window.PropertyHubCalendar = new PropertyHubCalendar();
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubCalendar;
}
