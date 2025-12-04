/**
 * PropertyHub Booking Management System
 * TREC-compliant booking system with calendar, payments, and compliance
 * Handles complete booking lifecycle from inquiry to completion
 */

class PropertyHubBooking {
    constructor() {
        this.currentBooking = null;
        this.availabilityCalendar = null;
        this.selectedDates = {};
        this.paymentProcessor = null;
        this.complianceChecks = {};
        this.bookingSteps = [];
        this.currentStep = 0;
        this.validationRules = {};
        
        // TREC compliance requirements
        this.trecRequirements = {
            requiredDisclosures: [
                'property_condition',
                'lead_paint',
                'environmental_hazards',
                'neighborhood_information',
                'seller_disclosures'
            ],
            requiredForms: [
                'residential_purchase_contract',
                'seller_disclosure',
                'lead_paint_disclosure',
                'property_condition_disclosure'
            ],
            mandatoryWaitingPeriods: {
                inspection: 7, // days
                financing: 21, // days
                appraisal: 14 // days
            }
        };

        // Booking states
        this.bookingStates = {
            INQUIRY: 'inquiry',
            PENDING: 'pending',
            CONFIRMED: 'confirmed',
            IN_PROGRESS: 'in_progress',
            COMPLETED: 'completed',
            CANCELLED: 'cancelled',
            DISPUTE: 'dispute'
        };

        this.init();
    }

    async init() {
        console.log('🏠 PropertyHub Booking System initializing...');
        
        await this.loadBookingData();
        this.setupBookingInterface();
        this.initializeCalendar();
        this.setupPaymentProcessor();
        this.setupComplianceValidation();
        this.setupEventListeners();
        
        console.log('✅ PropertyHub Booking System ready');
    }

    // Booking Process Management
    async startBookingProcess(propertyId, bookingType = 'purchase') {
        try {
            // Initialize new booking
            this.currentBooking = {
                id: this.generateBookingId(),
                propertyId: propertyId,
                type: bookingType,
                status: this.bookingStates.INQUIRY,
                createdAt: new Date().toISOString(),
                steps: this.getBookingSteps(bookingType),
                compliance: {
                    disclosures: {},
                    forms: {},
                    signatures: {},
                    deadlines: {}
                },
                timeline: [],
                participants: {
                    buyer: null,
                    seller: null,
                    buyerAgent: null,
                    sellerAgent: null,
                    titleCompany: null,
                    lender: null
                }
            };

            // Load property details
            const property = await this.loadPropertyDetails(propertyId);
            this.currentBooking.property = property;

            // Setup booking workflow
            this.setupBookingWorkflow();
            
            // Show booking interface
            this.showBookingInterface();
            
            // Track booking initiation
            this.trackBookingEvent('booking_initiated', {
                propertyId: propertyId,
                bookingType: bookingType
            });

        } catch (error) {
            console.error('Error starting booking process:', error);
            this.showBookingError('Failed to start booking process', error);
        }
    }

    getBookingSteps(bookingType) {
        const commonSteps = [
            {
                id: 'property_selection',
                title: 'Property Selection',
                description: 'Select and review property details',
                required: true,
                completed: false
            },
            {
                id: 'buyer_information',
                title: 'Buyer Information',
                description: 'Provide buyer contact and financial information',
                required: true,
                completed: false
            },
            {
                id: 'trec_disclosures',
                title: 'TREC Disclosures',
                description: 'Review and acknowledge required TREC disclosures',
                required: true,
                completed: false
            },
            {
                id: 'purchase_contract',
                title: 'Purchase Contract',
                description: 'Review and execute purchase agreement',
                required: true,
                completed: false
            },
            {
                id: 'earnest_money',
                title: 'Earnest Money',
                description: 'Submit earnest money deposit',
                required: true,
                completed: false
            },
            {
                id: 'inspection_period',
                title: 'Inspection Period',
                description: 'Property inspection and due diligence',
                required: true,
                completed: false,
                duration: this.trecRequirements.mandatoryWaitingPeriods.inspection
            },
            {
                id: 'financing',
                title: 'Financing',
                description: 'Secure mortgage financing',
                required: true,
                completed: false,
                duration: this.trecRequirements.mandatoryWaitingPeriods.financing
            },
            {
                id: 'appraisal',
                title: 'Appraisal',
                description: 'Property appraisal and valuation',
                required: true,
                completed: false,
                duration: this.trecRequirements.mandatoryWaitingPeriods.appraisal
            },
            {
                id: 'title_work',
                title: 'Title Work',
                description: 'Title search and insurance',
                required: true,
                completed: false
            },
            {
                id: 'final_walkthrough',
                title: 'Final Walkthrough',
                description: 'Pre-closing property inspection',
                required: true,
                completed: false
            },
            {
                id: 'closing',
                title: 'Closing',
                description: 'Final settlement and key transfer',
                required: true,
                completed: false
            }
        ];

        return commonSteps;
    }

    // TREC Compliance Management
    async loadTRECDisclosures(propertyId) {
        try {
            const response = await fetch(`/api/properties/${propertyId}/trec-disclosures`, {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to load TREC disclosures');
            }

            const disclosures = await response.json();
            return disclosures;
        } catch (error) {
            console.error('Error loading TREC disclosures:', error);
            throw error;
        }
    }

    renderTRECDisclosures(disclosures) {
        const container = document.getElementById('trec-disclosures');
        if (!container) return;

        container.innerHTML = `
            <div class="trec-disclosures-container">
                <div class="disclosure-header">
                    <h3>Required TREC Disclosures</h3>
                    <p class="disclaimer">
                        Texas Real Estate Commission (TREC) requires these disclosures for all real estate transactions.
                        Please review each disclosure carefully before proceeding.
                    </p>
                </div>

                ${disclosures.map(disclosure => `
                    <div class="disclosure-section" data-disclosure="${disclosure.id}">
                        <div class="disclosure-title">
                            <h4>${disclosure.title}</h4>
                            <span class="required-badge">Required by TREC</span>
                        </div>
                        
                        <div class="disclosure-content">
                            <div class="disclosure-text">
                                ${disclosure.content}
                            </div>
                            
                            ${disclosure.additionalDocuments ? `
                                <div class="additional-documents">
                                    <h5>Related Documents:</h5>
                                    <ul>
                                        ${disclosure.additionalDocuments.map(doc => `
                                            <li>
                                                <a href="${doc.url}" target="_blank" rel="noopener">
                                                    <i class="fas fa-file-pdf"></i> ${doc.title}
                                                </a>
                                            </li>
                                        `).join('')}
                                    </ul>
                                </div>
                            ` : ''}
                        </div>
                        
                        <div class="disclosure-acknowledgment">
                            <label class="checkbox-label">
                                <input type="checkbox" 
                                       name="disclosure_acknowledged" 
                                       value="${disclosure.id}"
                                       onchange="PropertyHubBooking.handleDisclosureAcknowledgment('${disclosure.id}', this.checked)">
                                I acknowledge that I have read and understand this disclosure
                            </label>
                            
                            <div class="signature-section" style="display: none;">
                                <label>Electronic Signature:</label>
                                <input type="text" 
                                       class="form-control signature-input" 
                                       placeholder="Type your full legal name"
                                       data-disclosure="${disclosure.id}">
                                <p class="signature-note">
                                    By typing your name above, you are providing your electronic signature 
                                    as acknowledgment of this disclosure.
                                </p>
                            </div>
                        </div>
                    </div>
                `).join('')}

                <div class="disclosure-completion">
                    <div class="progress-indicator">
                        <span id="disclosure-progress">0</span> of ${disclosures.length} disclosures acknowledged
                    </div>
                    
                    <button class="btn btn-primary btn-lg" 
                            id="continue-to-contract" 
                            disabled
                            onclick="PropertyHubBooking.proceedToContract()">
                        Continue to Purchase Contract
                    </button>
                </div>
            </div>
        `;

        this.updateDisclosureProgress();
    }

    handleDisclosureAcknowledgment(disclosureId, acknowledged) {
        const signatureSection = document.querySelector(`[data-disclosure="${disclosureId}"] .signature-section`);
        
        if (acknowledged) {
            signatureSection.style.display = 'block';
            this.currentBooking.compliance.disclosures[disclosureId] = {
                acknowledged: true,
                acknowledgedAt: new Date().toISOString()
            };
        } else {
            signatureSection.style.display = 'none';
            delete this.currentBooking.compliance.disclosures[disclosureId];
        }

        this.updateDisclosureProgress();
    }

    updateDisclosureProgress() {
        const acknowledgedCount = Object.keys(this.currentBooking.compliance.disclosures || {}).length;
        const totalDisclosures = this.trecRequirements.requiredDisclosures.length;
        
        const progressElement = document.getElementById('disclosure-progress');
        const continueButton = document.getElementById('continue-to-contract');
        
        if (progressElement) {
            progressElement.textContent = acknowledgedCount;
        }
        
        if (continueButton) {
            continueButton.disabled = acknowledgedCount < totalDisclosures;
        }
    }

    // Purchase Contract Management
    async generatePurchaseContract() {
        try {
            const response = await fetch('/api/bookings/generate-contract', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    bookingId: this.currentBooking.id,
                    propertyId: this.currentBooking.propertyId,
                    buyerInfo: this.currentBooking.buyer,
                    terms: this.currentBooking.terms
                })
            });

            if (!response.ok) {
                throw new Error('Failed to generate purchase contract');
            }

            const contract = await response.json();
            this.renderPurchaseContract(contract);
            
            return contract;
        } catch (error) {
            console.error('Error generating purchase contract:', error);
            throw error;
        }
    }

    renderPurchaseContract(contract) {
        const container = document.getElementById('purchase-contract');
        if (!container) return;

        container.innerHTML = `
            <div class="contract-container">
                <div class="contract-header">
                    <h3>Residential Purchase Contract</h3>
                    <div class="contract-info">
                        <p><strong>Contract #:</strong> ${contract.contractNumber}</p>
                        <p><strong>Generated:</strong> ${new Date(contract.generatedAt).toLocaleString()}</p>
                        <p><strong>TREC Form:</strong> ${contract.trecForm}</p>
                    </div>
                </div>

                <div class="contract-terms">
                    <div class="terms-section">
                        <h4>Property Information</h4>
                        <div class="terms-grid">
                            <div class="term-item">
                                <label>Property Address:</label>
                                <span>${contract.property.address}</span>
                            </div>
                            <div class="term-item">
                                <label>Legal Description:</label>
                                <span>${contract.property.legalDescription}</span>
                            </div>
                            <div class="term-item">
                                <label>MLS Number:</label>
                                <span>${contract.property.mlsNumber}</span>
                            </div>
                        </div>
                    </div>

                    <div class="terms-section">
                        <h4>Purchase Terms</h4>
                        <div class="terms-grid">
                            <div class="term-item">
                                <label>Purchase Price:</label>
                                <span class="price">$${contract.terms.purchasePrice.toLocaleString()}</span>
                            </div>
                            <div class="term-item">
                                <label>Earnest Money:</label>
                                <span>$${contract.terms.earnestMoney.toLocaleString()}</span>
                            </div>
                            <div class="term-item">
                                <label>Down Payment:</label>
                                <span>$${contract.terms.downPayment.toLocaleString()}</span>
                            </div>
                            <div class="term-item">
                                <label>Financing Amount:</label>
                                <span>$${contract.terms.financingAmount.toLocaleString()}</span>
                            </div>
                        </div>
                    </div>

                    <div class="terms-section">
                        <h4>Important Dates</h4>
                        <div class="terms-grid">
                            <div class="term-item">
                                <label>Contract Date:</label>
                                <span>${new Date(contract.dates.contractDate).toLocaleDateString()}</span>
                            </div>
                            <div class="term-item">
                                <label>Option Period Expires:</label>
                                <span>${new Date(contract.dates.optionExpiration).toLocaleDateString()}</span>
                            </div>
                            <div class="term-item">
                                <label>Financing Deadline:</label>
                                <span>${new Date(contract.dates.financingDeadline).toLocaleDateString()}</span>
                            </div>
                            <div class="term-item">
                                <label>Closing Date:</label>
                                <span>${new Date(contract.dates.closingDate).toLocaleDateString()}</span>
                            </div>
                        </div>
                    </div>

                    <div class="terms-section">
                        <h4>Contingencies</h4>
                        <ul class="contingencies-list">
                            ${contract.contingencies.map(contingency => `
                                <li>
                                    <strong>${contingency.title}:</strong> 
                                    ${contingency.description}
                                    <span class="contingency-deadline">
                                        (Deadline: ${new Date(contingency.deadline).toLocaleDateString()})
                                    </span>
                                </li>
                            `).join('')}
                        </ul>
                    </div>
                </div>

                <div class="contract-actions">
                    <div class="contract-warning">
                        <i class="fas fa-exclamation-triangle"></i>
                        <p>
                            This is a legally binding contract. Please review all terms carefully 
                            and consult with your real estate agent before signing.
                        </p>
                    </div>
                    
                    <div class="signature-section">
                        <h4>Electronic Signatures</h4>
                        
                        <div class="signature-party">
                            <h5>Buyer Signature</h5>
                            <input type="text" 
                                   class="form-control signature-input" 
                                   id="buyer-signature"
                                   placeholder="Type your full legal name">
                            <div class="signature-info">
                                <p>Date: ${new Date().toLocaleDateString()}</p>
                                <p>IP Address: <span id="user-ip">${contract.signingInfo.ipAddress}</span></p>
                            </div>
                        </div>
                        
                        <div class="contract-acknowledgments">
                            <label class="checkbox-label">
                                <input type="checkbox" id="terms-agreement" required>
                                I have read and agree to all terms and conditions in this contract
                            </label>
                            
                            <label class="checkbox-label">
                                <input type="checkbox" id="disclosure-acknowledgment" required>
                                I acknowledge receipt of all required TREC disclosures
                            </label>
                            
                            <label class="checkbox-label">
                                <input type="checkbox" id="electronic-signature-consent" required>
                                I consent to the use of electronic signatures for this transaction
                            </label>
                        </div>
                        
                        <button class="btn btn-success btn-lg" 
                                id="sign-contract" 
                                onclick="PropertyHubBooking.signContract()"
                                disabled>
                            <i class="fas fa-pen"></i> Sign Contract
                        </button>
                    </div>
                </div>

                <div class="contract-download">
                    <a href="${contract.pdfUrl}" 
                       target="_blank" 
                       class="btn btn-outline btn-sm">
                        <i class="fas fa-download"></i> Download PDF Copy
                    </a>
                </div>
            </div>
        `;

        this.setupContractValidation();
    }

    // Payment Processing
    async initializePaymentProcessor() {
        // Initialize payment processor (Stripe, Square, etc.)
        try {
            if (typeof Stripe !== 'undefined') {
                this.paymentProcessor = Stripe(window.STRIPE_PUBLIC_KEY);
            }
        } catch (error) {
            console.error('Error initializing payment processor:', error);
        }
    }

    async processEarnestMoney(amount, paymentMethod) {
        try {
            const response = await fetch('/api/payments/earnest-money', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    bookingId: this.currentBooking.id,
                    amount: amount,
                    paymentMethod: paymentMethod,
                    description: `Earnest money for ${this.currentBooking.property.address}`
                })
            });

            if (!response.ok) {
                throw new Error('Payment processing failed');
            }

            const payment = await response.json();
            
            // Update booking status
            this.currentBooking.payments = this.currentBooking.payments || [];
            this.currentBooking.payments.push(payment);
            
            // Track payment event
            this.trackBookingEvent('earnest_money_paid', {
                amount: amount,
                paymentId: payment.id
            });

            this.showPaymentSuccess(payment);
            this.proceedToNextStep();
            
            return payment;
        } catch (error) {
            console.error('Error processing earnest money:', error);
            this.showPaymentError(error.message);
            throw error;
        }
    }

    // Calendar and Scheduling
    async loadAvailabilityCalendar(propertyId) {
        try {
            const response = await fetch(`/api/properties/${propertyId}/availability`, {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            const availability = await response.json();
            this.renderAvailabilityCalendar(availability);
            
            return availability;
        } catch (error) {
            console.error('Error loading availability calendar:', error);
        }
    }

    renderAvailabilityCalendar(availability) {
        const container = document.getElementById('availability-calendar');
        if (!container) return;

        // This would integrate with a calendar library like FullCalendar
        container.innerHTML = `
            <div class="calendar-container">
                <div class="calendar-header">
                    <h4>Schedule Inspection</h4>
                    <p>Select available dates for property inspection</p>
                </div>
                
                <div id="inspection-calendar" class="inspection-calendar"></div>
                
                <div class="time-slots" id="time-slots" style="display: none;">
                    <h5>Available Times</h5>
                    <div class="time-slot-grid">
                        <!-- Time slots will be populated dynamically -->
                    </div>
                </div>
                
                <div class="calendar-actions">
                    <button class="btn btn-primary" 
                            id="schedule-inspection"
                            onclick="PropertyHubBooking.scheduleInspection()"
                            disabled>
                        Schedule Inspection
                    </button>
                </div>
            </div>
        `;

        // Initialize calendar widget (would use FullCalendar or similar)
        this.initializeCalendarWidget(availability);
    }

    // Workflow Management
    async proceedToNextStep() {
        const currentStepIndex = this.currentStep;
        const nextStepIndex = currentStepIndex + 1;
        
        if (nextStepIndex < this.currentBooking.steps.length) {
            // Mark current step as completed
            this.currentBooking.steps[currentStepIndex].completed = true;
            this.currentBooking.steps[currentStepIndex].completedAt = new Date().toISOString();
            
            // Move to next step
            this.currentStep = nextStepIndex;
            
            // Update UI
            this.updateBookingProgress();
            this.showCurrentStep();
            
            // Save progress
            await this.saveBookingProgress();
            
            // Check if booking is complete
            if (this.isBookingComplete()) {
                await this.completeBooking();
            }
        }
    }

    updateBookingProgress() {
        const progressContainer = document.getElementById('booking-progress');
        if (!progressContainer) return;

        const completedSteps = this.currentBooking.steps.filter(step => step.completed).length;
        const totalSteps = this.currentBooking.steps.length;
        const progressPercentage = (completedSteps / totalSteps) * 100;

        progressContainer.innerHTML = `
            <div class="progress-header">
                <h4>Booking Progress</h4>
                <span class="progress-text">${completedSteps} of ${totalSteps} steps completed</span>
            </div>
            
            <div class="progress-bar-container">
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${progressPercentage}%"></div>
                </div>
                <span class="progress-percentage">${Math.round(progressPercentage)}%</span>
            </div>
            
            <div class="steps-list">
                ${this.currentBooking.steps.map((step, index) => `
                    <div class="step-item ${step.completed ? 'completed' : ''} ${index === this.currentStep ? 'current' : ''}">
                        <div class="step-icon">
                            ${step.completed ? '<i class="fas fa-check"></i>' : (index === this.currentStep ? '<i class="fas fa-circle"></i>' : '<i class="far fa-circle"></i>')}
                        </div>
                        <div class="step-content">
                            <h5>${step.title}</h5>
                            <p>${step.description}</p>
                            ${step.duration ? `<span class="step-duration">${step.duration} days</span>` : ''}
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    // Event Listeners
    setupEventListeners() {
        // Contract validation
        document.addEventListener('change', (e) => {
            if (e.target.matches('#terms-agreement, #disclosure-acknowledgment, #electronic-signature-consent')) {
                this.validateContractSignature();
            }
        });

        // Signature input validation
        document.addEventListener('input', (e) => {
            if (e.target.matches('#buyer-signature')) {
                this.validateContractSignature();
            }
        });

        // Date selection
        document.addEventListener('click', (e) => {
            if (e.target.matches('.calendar-date')) {
                this.selectInspectionDate(e.target.dataset.date);
            }
        });

        // Form submissions
        document.addEventListener('submit', (e) => {
            if (e.target.matches('.booking-form')) {
                e.preventDefault();
                this.handleBookingFormSubmission(e.target);
            }
        });
    }

    setupContractValidation() {
        const signatureInput = document.getElementById('buyer-signature');
        const checkboxes = document.querySelectorAll('#terms-agreement, #disclosure-acknowledgment, #electronic-signature-consent');
        const signButton = document.getElementById('sign-contract');

        const validateSignature = () => {
            const hasSignature = signatureInput && signatureInput.value.trim().length >= 3;
            const allChecked = Array.from(checkboxes).every(checkbox => checkbox.checked);
            
            if (signButton) {
                signButton.disabled = !hasSignature || !allChecked;
            }
        };

        if (signatureInput) {
            signatureInput.addEventListener('input', validateSignature);
        }

        checkboxes.forEach(checkbox => {
            checkbox.addEventListener('change', validateSignature);
        });
    }

    // Booking Interface
    setupBookingInterface() {
        const bookingContainer = document.getElementById('booking-container');
        if (!bookingContainer) return;

        bookingContainer.innerHTML = `
            <div class="booking-interface">
                <div class="booking-sidebar">
                    <div class="property-summary" id="property-summary"></div>
                    <div class="booking-progress" id="booking-progress"></div>
                    <div class="booking-timeline" id="booking-timeline"></div>
                </div>
                
                <div class="booking-main">
                    <div class="booking-step-content" id="booking-step-content"></div>
                </div>
            </div>
        `;
    }

    showCurrentStep() {
        const currentStepData = this.currentBooking.steps[this.currentStep];
        const contentContainer = document.getElementById('booking-step-content');
        
        if (!contentContainer || !currentStepData) return;

        // Show step-specific content
        switch (currentStepData.id) {
            case 'property_selection':
                this.showPropertySelection();
                break;
            case 'buyer_information':
                this.showBuyerInformationForm();
                break;
            case 'trec_disclosures':
                this.showTRECDisclosures();
                break;
            case 'purchase_contract':
                this.showPurchaseContract();
                break;
            case 'earnest_money':
                this.showEarnestMoneyPayment();
                break;
            case 'inspection_period':
                this.showInspectionScheduling();
                break;
            case 'financing':
                this.showFinancingStatus();
                break;
            case 'closing':
                this.showClosingPreparation();
                break;
        }
    }

    // Utility Methods
    generateBookingId() {
        return 'booking_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    trackBookingEvent(eventName, data = {}) {
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackEvent(eventName, {
                bookingId: this.currentBooking?.id,
                propertyId: this.currentBooking?.propertyId,
                step: this.currentBooking?.steps[this.currentStep]?.id,
                ...data
            });
        }
    }

    showBookingError(message, error) {
        console.error(message, error);
        const container = document.getElementById('booking-step-content');
        if (container) {
            container.innerHTML = `
                <div class="booking-error">
                    <i class="fas fa-exclamation-triangle fa-2x"></i>
                    <h3>Booking Error</h3>
                    <p>${message}</p>
                    <button class="btn btn-primary" onclick="location.reload()">
                        Retry
                    </button>
                </div>
            `;
        }
    }

    showPaymentSuccess(payment) {
        const notification = {
            type: 'success',
            title: 'Payment Successful',
            message: `Earnest money payment of $${payment.amount.toLocaleString()} has been processed successfully.`
        };

        if (window.PropertyHubDashboard) {
            window.PropertyHubDashboard.displayNotification(notification);
        }
    }

    showPaymentError(message) {
        const notification = {
            type: 'error',
            title: 'Payment Failed',
            message: message
        };

        if (window.PropertyHubDashboard) {
            window.PropertyHubDashboard.displayNotification(notification);
        }
    }

    // Public API Methods
    async signContract() {
        const signature = document.getElementById('buyer-signature')?.value;
        
        if (!signature || signature.trim().length < 3) {
            alert('Please provide your full legal name as electronic signature');
            return;
        }

        try {
            const response = await fetch('/api/bookings/sign-contract', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    bookingId: this.currentBooking.id,
                    signature: signature,
                    timestamp: new Date().toISOString(),
                    ipAddress: await this.getClientIP()
                })
            });

            if (!response.ok) {
                throw new Error('Contract signing failed');
            }

            const result = await response.json();
            
            // Update booking status
            this.currentBooking.compliance.signatures.buyer = {
                signature: signature,
                signedAt: new Date().toISOString(),
                ipAddress: result.ipAddress
            };

            this.trackBookingEvent('contract_signed');
            this.proceedToNextStep();
            
        } catch (error) {
            console.error('Error signing contract:', error);
            alert('Failed to sign contract. Please try again.');
        }
    }

    async scheduleInspection() {
        // Implementation for scheduling inspection
        console.log('Scheduling inspection...');
    }

    async proceedToContract() {
        // Validate all disclosures are acknowledged
        const requiredDisclosures = this.trecRequirements.requiredDisclosures.length;
        const acknowledgedDisclosures = Object.keys(this.currentBooking.compliance.disclosures || {}).length;
        
        if (acknowledgedDisclosures < requiredDisclosures) {
            alert('Please acknowledge all required TREC disclosures before proceeding.');
            return;
        }

        this.proceedToNextStep();
    }

    validateContractSignature() {
        const signatureInput = document.getElementById('buyer-signature');
        const checkboxes = document.querySelectorAll('#terms-agreement, #disclosure-acknowledgment, #electronic-signature-consent');
        const signButton = document.getElementById('sign-contract');

        if (!signatureInput || !signButton) return;

        const hasSignature = signatureInput.value.trim().length >= 3;
        const allChecked = Array.from(checkboxes).every(checkbox => checkbox.checked);
        
        signButton.disabled = !hasSignature || !allChecked;
    }

    async getClientIP() {
        try {
            const response = await fetch('/api/utils/client-ip');
            const data = await response.json();
            return data.ip;
        } catch (error) {
            return 'unknown';
        }
    }

    async loadBookingData() {
        // Load any existing booking data
        const bookingId = new URLSearchParams(window.location.search).get('bookingId');
        if (bookingId) {
            try {
                const response = await fetch(`/api/bookings/${bookingId}`, {
                    headers: {
                        'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                    }
                });

                if (response.ok) {
                    this.currentBooking = await response.json();
                }
            } catch (error) {
                console.error('Error loading booking data:', error);
            }
        }
    }

    async saveBookingProgress() {
        if (!this.currentBooking) return;

        try {
            await fetch(`/api/bookings/${this.currentBooking.id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify(this.currentBooking)
            });
        } catch (error) {
            console.error('Error saving booking progress:', error);
        }
    }

    isBookingComplete() {
        return this.currentBooking.steps.every(step => step.completed);
    }

    async completeBooking() {
        this.currentBooking.status = this.bookingStates.COMPLETED;
        this.currentBooking.completedAt = new Date().toISOString();
        
        await this.saveBookingProgress();
        this.trackBookingEvent('booking_completed');
        
        // Show completion message
        this.showBookingCompletion();
    }
}

// Initialize PropertyHub Booking when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    if (document.querySelector('#booking-container') || window.location.pathname.includes('/booking')) {
        window.PropertyHubBooking = new PropertyHubBooking();
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubBooking;
}
