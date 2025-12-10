function bookingCalendar() {
    return {
        selectedDate: null,
        selectedTime: null,
        availableSlots: [],
        propertyId: null,
        loading: false,
        error: null,
        
        init() {
            const params = new URLSearchParams(window.location.search);
            this.propertyId = params.get('property_id');
            if (this.propertyId) {
                this.loadAvailability();
            }
        },
        
        async loadAvailability() {
            this.loading = true;
            this.error = null;
            try {
                const response = await fetch(`/api/v1/availability/check?property_id=${this.propertyId}`);
                const data = await response.json();
                if (data.success) {
                    this.availableSlots = data.data.available_slots || [];
                } else {
                    this.error = data.error || 'Failed to load availability';
                }
            } catch (err) {
                Logger.error('Failed to load availability:', err);
                this.error = 'Network error loading availability';
            } finally {
                this.loading = false;
            }
        },
        
        selectDate(date) {
            this.selectedDate = date;
            this.selectedTime = null;
        },
        
        selectTime(time) {
            this.selectedTime = time;
        },
        
        isDateAvailable(date) {
            if (!this.availableSlots || this.availableSlots.length === 0) {
                return true;
            }
            
            const dateStr = typeof date === 'string' ? date : date.toISOString().split('T')[0];
            return this.availableSlots.some(slot => {
                const slotDate = new Date(slot.date).toISOString().split('T')[0];
                return slotDate === dateStr && !slot.is_blackout;
            });
        },
        
        getAvailableTimesForDate(date) {
            if (!this.availableSlots || this.availableSlots.length === 0) {
                return [];
            }
            
            const dateStr = typeof date === 'string' ? date : date.toISOString().split('T')[0];
            const slots = this.availableSlots.filter(slot => {
                const slotDate = new Date(slot.date).toISOString().split('T')[0];
                return slotDate === dateStr && !slot.is_blackout;
            });
            
            return slots.map(slot => slot.time);
        }
    };
}

function bookingForm() {
    return {
        step: 1,
        loading: false,
        error: null,
        success: false,
        
        form: {
            propertyId: new URLSearchParams(window.location.search).get('property_id') || '',
            firstName: '',
            lastName: '',
            email: '',
            phone: '',
            showingDate: '',
            showingTime: '',
            showingType: 'in-person',
            attendeeCount: 1,
            hasAgent: 'no',
            notes: '',
            consentGiven: false,
            marketingOptIn: false
        },
        
        init() {
            const urlPropertyId = new URLSearchParams(window.location.search).get('property_id');
            if (urlPropertyId) {
                this.form.propertyId = urlPropertyId;
                const selectEl = document.getElementById('propertySelect');
                if (selectEl) {
                    selectEl.value = urlPropertyId;
                }
            }
        },
        
        errors: {},
        requirementsChecked: {
            income: false,
            felony: false,
            eviction: false,
            credit: false
        },
        
        get checkedCount() {
            return Object.values(this.requirementsChecked).filter(v => v).length;
        },
        
        get allRequirementsMet() {
            return this.requirementsChecked.income && 
                   this.requirementsChecked.felony && 
                   this.requirementsChecked.eviction && 
                   this.requirementsChecked.credit;
        },
        
        get progressMessage() {
            const count = this.checkedCount;
            if (count === 0) {
                return 'Tap each box to confirm you meet the requirement';
            } else if (count === 4) {
                return "Great! You're ready to schedule your tour";
            } else {
                const remaining = 4 - count;
                return `Almost there! ${remaining} more to go`;
            }
        },
        
        get canProceedToStep2() {
            return this.allRequirementsMet;
        },
        
        get canProceedToStep3() {
            return this.form.propertyId && this.form.showingDate && this.form.showingTime;
        },
        
        nextStep() {
            if (this.validateCurrentStep()) {
                this.step++;
                window.scrollTo(0, 0);
            }
        },
        
        prevStep() {
            if (this.step > 1) {
                this.step--;
                window.scrollTo(0, 0);
            }
        },
        
        validateCurrentStep() {
            this.errors = {};
            
            switch(this.step) {
                case 1:
                    if (!this.allRequirementsMet) {
                        this.errors.requirements = 'Please confirm all requirements';
                        return false;
                    }
                    break;
                    
                case 2:
                    if (!this.form.propertyId) {
                        this.errors.propertyId = 'Please select a property';
                    }
                    if (!this.form.showingDate) {
                        this.errors.showingDate = 'Please select a date';
                    }
                    if (!this.form.showingTime) {
                        this.errors.showingTime = 'Please select a time';
                    }
                    break;
                    
                case 3:
                    if (!this.form.firstName || this.form.firstName.trim().length < 2) {
                        this.errors.firstName = 'First name is required';
                    }
                    if (!this.form.lastName || this.form.lastName.trim().length < 2) {
                        this.errors.lastName = 'Last name is required';
                    }
                    if (!this.form.email || !this.isValidEmail(this.form.email)) {
                        this.errors.email = 'Valid email is required';
                    }
                    if (!this.form.phone || this.form.phone.trim().length < 10) {
                        this.errors.phone = 'Valid phone number is required';
                    }
                    if (!this.form.consentGiven) {
                        this.errors.consentGiven = 'You must agree to the terms';
                    }
                    break;
            }
            
            return Object.keys(this.errors).length === 0;
        },
        
        isValidEmail(email) {
            const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return re.test(email);
        },
        
        async submitBooking() {
            if (!this.validateCurrentStep()) {
                return;
            }
            
            this.loading = true;
            this.error = null;
            
            try {
                const response = await fetch('/api/v1/bookings', {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': document.querySelector('input[name="csrf_token"]')?.value || ''
                    },
                    body: JSON.stringify({
                        property_id: this.form.propertyId,
                        first_name: this.form.firstName,
                        last_name: this.form.lastName,
                        email: this.form.email,
                        phone: this.form.phone,
                        showing_date: this.form.showingDate,
                        showing_time: this.form.showingTime,
                        showing_type: this.form.showingType,
                        attendee_count: this.form.attendeeCount,
                        has_agent: this.form.hasAgent === 'yes',
                        notes: this.form.notes,
                        consent_given: this.form.consentGiven,
                        marketing_opt_in: this.form.marketingOptIn
                    })
                });
                
                const data = await response.json();
                
                if (!response.ok) {
                    if (response.status === 400) {
                        this.error = data.message || data.error || 'Invalid booking data';
                        if (data.errors) {
                            this.errors = data.errors;
                        }
                    } else if (response.status === 404) {
                        this.error = 'Property not found. Please select a valid property.';
                    } else if (response.status === 409) {
                        this.error = 'This time slot is no longer available. Please select another.';
                        if (data.alternative_slots) {
                            this.error += ' Alternative slots are available.';
                        }
                    } else if (response.status === 429) {
                        this.error = 'Too many booking attempts. Please wait a few minutes.';
                    } else {
                        this.error = data.error || 'Failed to create booking. Please try again.';
                    }
                    return;
                }
                
                if (data.success) {
                    this.success = true;
                    window.location.href = `/booking-confirmed?ref=${data.data.reference_number}`;
                } else {
                    this.error = data.error || 'Failed to create booking';
                    if (data.errors) {
                        this.errors = data.errors;
                    }
                }
            } catch (err) {
                Logger.error('Booking submission error:', err);
                this.error = 'Network error. Please try again.';
            } finally {
                this.loading = false;
            }
        }
    };
}

function bookingConfirmation() {
    return {
        booking: null,
        loading: true,
        error: null,
        
        async init() {
            const params = new URLSearchParams(window.location.search);
            const ref = params.get('ref');
            
            if (!ref) {
                this.error = 'No booking reference provided';
                this.loading = false;
                return;
            }
            
            try {
                const response = await fetch(`/api/v1/bookings?reference=${ref}`);
                const data = await response.json();
                
                if (data.success && data.data) {
                    this.booking = data.data;
                } else {
                    this.error = 'Booking not found';
                }
            } catch (err) {
                Logger.error('Failed to load booking:', err);
                this.error = 'Failed to load booking details';
            } finally {
                this.loading = false;
            }
        },
        
        formatDate(dateStr) {
            if (!dateStr) return '';
            const date = new Date(dateStr);
            return date.toLocaleDateString('en-US', { 
                weekday: 'long', 
                year: 'numeric', 
                month: 'long', 
                day: 'numeric' 
            });
        },
        
        formatTime(timeStr) {
            if (!timeStr) return '';
            return timeStr;
        }
    };
}

if (typeof window !== 'undefined') {
    window.bookingCalendar = bookingCalendar;
    window.bookingForm = bookingForm;
    window.bookingConfirmation = bookingConfirmation;
}
