/**
 * PropertyHub Form Validation System
 * Advanced form validation with real estate-specific rules
 * TREC compliance, financial validation, and user experience optimization
 */

class PropertyHubFormValidation {
    constructor() {
        this.validators = new Map();
        this.validationRules = new Map();
        this.errorMessages = new Map();
        this.validationState = new Map();
        this.asyncValidators = new Map();
        this.validationGroups = new Map();
        
        // Real estate specific validation patterns
        this.patterns = {
            mlsNumber: /^[A-Z0-9]{6,12}$/,
            zipCode: /^\d{5}(-\d{4})?$/,
            phone: /^\(?([0-9]{3})\)?[-. ]?([0-9]{3})[-. ]?([0-9]{4})$/,
            email: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
            ssn: /^\d{3}-?\d{2}-?\d{4}$/,
            currency: /^\$?([0-9]{1,3},([0-9]{3},)*[0-9]{3}|[0-9]+)(\.[0-9][0-9])?$/,
            percentage: /^(100(\.0{1,2})?|[1-9]?\d(\.\d{1,2})?)$/,
            coordinates: /^-?([1-8]?\d(\.\d+)?|90(\.0+)?),\s*-?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$/
        };

        // TREC compliance requirements
        this.trecValidation = {
            requiredDisclosures: [
                'property_condition',
                'lead_paint',
                'environmental_hazards',
                'seller_disclosures'
            ],
            requiredSignatures: [
                'buyer_signature',
                'seller_signature',
                'agent_signature'
            ],
            dateValidations: {
                contractDate: { required: true, notPastDate: true },
                closingDate: { required: true, minDaysFromNow: 7 },
                inspectionDeadline: { required: true, maxDaysFromContract: 10 }
            }
        };

        // Financial validation rules
        this.financialValidation = {
            income: { min: 0, max: 10000000 },
            downPayment: { min: 0, max: 50000000 },
            loanAmount: { min: 1000, max: 50000000 },
            creditScore: { min: 300, max: 850 },
            debtToIncomeRatio: { min: 0, max: 100 }
        };

        this.init();
    }

    async init() {
        console.log('✅ PropertyHub Form Validation initializing...');
        
        this.setupValidationRules();
        this.setupErrorMessages();
        this.setupAsyncValidators();
        this.scanAndSetupForms();
        this.setupEventListeners();
        
        console.log('✅ PropertyHub Form Validation ready');
    }

    // Validation Rules Setup
    setupValidationRules() {
        // Basic validation rules
        this.validationRules.set('required', (value) => {
            return value !== null && value !== undefined && String(value).trim() !== '';
        });

        this.validationRules.set('email', (value) => {
            return !value || this.patterns.email.test(value);
        });

        this.validationRules.set('phone', (value) => {
            return !value || this.patterns.phone.test(value);
        });

        this.validationRules.set('zipCode', (value) => {
            return !value || this.patterns.zipCode.test(value);
        });

        this.validationRules.set('mlsNumber', (value) => {
            return !value || this.patterns.mlsNumber.test(value);
        });

        this.validationRules.set('ssn', (value) => {
            return !value || this.patterns.ssn.test(value);
        });

        this.validationRules.set('currency', (value) => {
            return !value || this.patterns.currency.test(value) || !isNaN(parseFloat(value));
        });

        // Numeric validations
        this.validationRules.set('min', (value, min) => {
            return !value || parseFloat(value) >= parseFloat(min);
        });

        this.validationRules.set('max', (value, max) => {
            return !value || parseFloat(value) <= parseFloat(max);
        });

        this.validationRules.set('minLength', (value, minLength) => {
            return !value || String(value).length >= parseInt(minLength);
        });

        this.validationRules.set('maxLength', (value, maxLength) => {
            return !value || String(value).length <= parseInt(maxLength);
        });

        // Date validations
        this.validationRules.set('futureDate', (value) => {
            if (!value) return true;
            const inputDate = new Date(value);
            const today = new Date();
            today.setHours(0, 0, 0, 0);
            return inputDate >= today;
        });

        this.validationRules.set('pastDate', (value) => {
            if (!value) return true;
            const inputDate = new Date(value);
            const today = new Date();
            return inputDate <= today;
        });

        this.validationRules.set('minAge', (value, minAge) => {
            if (!value) return true;
            const birthDate = new Date(value);
            const today = new Date();
            const age = today.getFullYear() - birthDate.getFullYear();
            const monthDiff = today.getMonth() - birthDate.getMonth();
            
            if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
                age--;
            }
            
            return age >= parseInt(minAge);
        });

        // Real estate specific validations
        this.validationRules.set('propertyPrice', (value) => {
            const price = parseFloat(value);
            return !value || (price >= 1000 && price <= 100000000);
        });

        this.validationRules.set('squareFootage', (value) => {
            const sqft = parseFloat(value);
            return !value || (sqft >= 100 && sqft <= 50000);
        });

        this.validationRules.set('lotSize', (value) => {
            const acres = parseFloat(value);
            return !value || (acres >= 0.01 && acres <= 1000);
        });

        this.validationRules.set('bedrooms', (value) => {
            const beds = parseInt(value);
            return !value || (beds >= 0 && beds <= 20);
        });

        this.validationRules.set('bathrooms', (value) => {
            const baths = parseFloat(value);
            return !value || (baths >= 0 && baths <= 20);
        });

        // Financial validations
        this.validationRules.set('creditScore', (value) => {
            const score = parseInt(value);
            return !value || (score >= 300 && score <= 850);
        });

        this.validationRules.set('annualIncome', (value) => {
            const income = parseFloat(value);
            return !value || (income >= 0 && income <= 10000000);
        });

        this.validationRules.set('downPaymentPercentage', (value) => {
            const percentage = parseFloat(value);
            return !value || (percentage >= 0 && percentage <= 100);
        });
    }

    setupErrorMessages() {
        this.errorMessages.set('required', 'This field is required');
        this.errorMessages.set('email', 'Please enter a valid email address');
        this.errorMessages.set('phone', 'Please enter a valid phone number (e.g., 555-123-4567)');
        this.errorMessages.set('zipCode', 'Please enter a valid ZIP code');
        this.errorMessages.set('mlsNumber', 'Please enter a valid MLS number (6-12 characters)');
        this.errorMessages.set('ssn', 'Please enter a valid Social Security Number');
        this.errorMessages.set('currency', 'Please enter a valid dollar amount');
        this.errorMessages.set('min', (min) => `Value must be at least ${min}`);
        this.errorMessages.set('max', (max) => `Value must be no more than ${max}`);
        this.errorMessages.set('minLength', (minLength) => `Must be at least ${minLength} characters`);
        this.errorMessages.set('maxLength', (maxLength) => `Must be no more than ${maxLength} characters`);
        this.errorMessages.set('futureDate', 'Date must be in the future');
        this.errorMessages.set('pastDate', 'Date must be in the past');
        this.errorMessages.set('minAge', (minAge) => `Must be at least ${minAge} years old`);
        this.errorMessages.set('propertyPrice', 'Property price must be between $1,000 and $100,000,000');
        this.errorMessages.set('squareFootage', 'Square footage must be between 100 and 50,000');
        this.errorMessages.set('lotSize', 'Lot size must be between 0.01 and 1,000 acres');
        this.errorMessages.set('bedrooms', 'Number of bedrooms must be between 0 and 20');
        this.errorMessages.set('bathrooms', 'Number of bathrooms must be between 0 and 20');
        this.errorMessages.set('creditScore', 'Credit score must be between 300 and 850');
        this.errorMessages.set('annualIncome', 'Annual income must be between $0 and $10,000,000');
        this.errorMessages.set('downPaymentPercentage', 'Down payment percentage must be between 0% and 100%');
        this.errorMessages.set('passwordMatch', 'Passwords do not match');
        this.errorMessages.set('uniqueEmail', 'This email address is already registered');
        this.errorMessages.set('uniqueUsername', 'This username is already taken');
        this.errorMessages.set('trecCompliance', 'TREC compliance requirements not met');
    }

    setupAsyncValidators() {
        // Email uniqueness check
        this.asyncValidators.set('uniqueEmail', async (value) => {
            if (!value || !this.patterns.email.test(value)) return true;
            
            try {
                const response = await fetch('/api/auth/check-email', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email: value })
                });
                
                const result = await response.json();
                return result.available;
            } catch (error) {
                console.error('Email validation error:', error);
                return true; // Don't fail on network errors
            }
        });

        // Username uniqueness check
        this.asyncValidators.set('uniqueUsername', async (value) => {
            if (!value || value.length < 3) return true;
            
            try {
                const response = await fetch('/api/auth/check-username', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username: value })
                });
                
                const result = await response.json();
                return result.available;
            } catch (error) {
                console.error('Username validation error:', error);
                return true;
            }
        });

        // MLS number validation
        this.asyncValidators.set('validMLS', async (value) => {
            if (!value || !this.patterns.mlsNumber.test(value)) return true;
            
            try {
                const response = await fetch('/api/properties/validate-mls', {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                    },
                    body: JSON.stringify({ mlsNumber: value })
                });
                
                const result = await response.json();
                return result.valid;
            } catch (error) {
                console.error('MLS validation error:', error);
                return true;
            }
        });

        // Address validation
        this.asyncValidators.set('validAddress', async (value) => {
            if (!value || value.length < 10) return true;
            
            try {
                const response = await fetch('/api/utils/validate-address', {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                    },
                    body: JSON.stringify({ address: value })
                });
                
                const result = await response.json();
                return result.valid;
            } catch (error) {
                console.error('Address validation error:', error);
                return true;
            }
        });
    }

    // Form Setup and Scanning
    scanAndSetupForms() {
        const forms = document.querySelectorAll('form[data-validate="true"], .validate-form');
        
        forms.forEach(form => {
            this.setupFormValidation(form);
        });
    }

    setupFormValidation(form) {
        const formId = form.id || this.generateFormId();
        const validationGroup = form.dataset.validationGroup || 'default';
        
        // Initialize form state
        this.validationState.set(formId, {
            isValid: false,
            errors: new Map(),
            touched: new Set(),
            submitted: false,
            asyncValidating: new Set()
        });

        // Setup field validation
        const fields = form.querySelectorAll('input, select, textarea');
        fields.forEach(field => {
            this.setupFieldValidation(formId, field);
        });

        // Setup form submission
        form.addEventListener('submit', (e) => {
            this.handleFormSubmission(e, formId);
        });

        // Add validation group
        this.validationGroups.set(formId, validationGroup);
        
        console.log(`📝 Form validation setup for: ${formId}`);
    }

    setupFieldValidation(formId, field) {
        const fieldName = field.name || field.id;
        if (!fieldName) return;

        // Parse validation rules from data attributes
        const rules = this.parseValidationRules(field);
        
        // Store rules for this field
        const fieldKey = `${formId}_${fieldName}`;
        this.validators.set(fieldKey, rules);

        // Setup event listeners
        this.setupFieldEventListeners(formId, field, fieldName);
        
        // Create error display element
        this.createErrorElement(field);
    }

    parseValidationRules(field) {
        const rules = [];
        
        // HTML5 validation attributes
        if (field.required) {
            rules.push({ type: 'required' });
        }
        
        if (field.type === 'email') {
            rules.push({ type: 'email' });
        }
        
        if (field.min) {
            rules.push({ type: 'min', value: field.min });
        }
        
        if (field.max) {
            rules.push({ type: 'max', value: field.max });
        }
        
        if (field.minLength) {
            rules.push({ type: 'minLength', value: field.minLength });
        }
        
        if (field.maxLength) {
            rules.push({ type: 'maxLength', value: field.maxLength });
        }

        // Custom validation attributes
        Object.keys(field.dataset).forEach(key => {
            if (key.startsWith('validate')) {
                const ruleType = key.replace('validate', '').toLowerCase();
                const ruleValue = field.dataset[key];
                
                if (this.validationRules.has(ruleType)) {
                    rules.push({ type: ruleType, value: ruleValue });
                }
            }
        });

        // Real estate specific rules based on field name/class
        if (field.classList.contains('mls-number') || field.name === 'mlsNumber') {
            rules.push({ type: 'mlsNumber' });
        }
        
        if (field.classList.contains('property-price') || field.name === 'price') {
            rules.push({ type: 'propertyPrice' });
        }
        
        if (field.classList.contains('zip-code') || field.name === 'zipCode') {
            rules.push({ type: 'zipCode' });
        }

        // Financial validations
        if (field.classList.contains('credit-score') || field.name === 'creditScore') {
            rules.push({ type: 'creditScore' });
        }

        if (field.classList.contains('annual-income') || field.name === 'annualIncome') {
            rules.push({ type: 'annualIncome' });
        }

        // Password confirmation
        if (field.type === 'password' && field.name === 'confirmPassword') {
            const passwordField = field.form.querySelector('input[name="password"]');
            if (passwordField) {
                rules.push({ 
                    type: 'custom', 
                    validator: (value) => value === passwordField.value,
                    message: 'Passwords do not match'
                });
            }
        }

        return rules;
    }

    setupFieldEventListeners(formId, field, fieldName) {
        // Real-time validation on input
        let timeout;
        field.addEventListener('input', (e) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => {
                this.validateField(formId, fieldName, e.target.value);
            }, 300);
        });

        // Validate on blur
        field.addEventListener('blur', (e) => {
            this.markFieldAsTouched(formId, fieldName);
            this.validateField(formId, fieldName, e.target.value);
        });

        // Clear validation on focus
        field.addEventListener('focus', (e) => {
            this.clearFieldError(formId, fieldName);
        });

        // Special handling for currency fields
        if (field.classList.contains('currency-input')) {
            field.addEventListener('input', (e) => {
                this.formatCurrencyInput(e.target);
            });
        }

        // Special handling for phone fields
        if (field.classList.contains('phone-input')) {
            field.addEventListener('input', (e) => {
                this.formatPhoneInput(e.target);
            });
        }
    }

    // Field Validation
    async validateField(formId, fieldName, value) {
        const fieldKey = `${formId}_${fieldName}`;
        const rules = this.validators.get(fieldKey);
        const formState = this.validationState.get(formId);
        
        if (!rules || !formState) return true;

        const errors = [];

        // Run synchronous validations
        for (const rule of rules) {
            const isValid = await this.executeValidationRule(rule, value);
            
            if (!isValid) {
                const errorMessage = this.getErrorMessage(rule.type, rule.value, rule.message);
                errors.push(errorMessage);
            }
        }

        // Run async validations
        const asyncRules = rules.filter(rule => this.asyncValidators.has(rule.type));
        if (asyncRules.length > 0) {
            formState.asyncValidating.add(fieldName);
            this.showFieldValidating(formId, fieldName);

            for (const rule of asyncRules) {
                try {
                    const validator = this.asyncValidators.get(rule.type);
                    const isValid = await validator(value);
                    
                    if (!isValid) {
                        const errorMessage = this.getErrorMessage(rule.type, rule.value, rule.message);
                        errors.push(errorMessage);
                    }
                } catch (error) {
                    console.error(`Async validation error for ${rule.type}:`, error);
                }
            }

            formState.asyncValidating.delete(fieldName);
            this.hideFieldValidating(formId, fieldName);
        }

        // Update form state
        if (errors.length > 0) {
            formState.errors.set(fieldName, errors);
            this.showFieldErrors(formId, fieldName, errors);
        } else {
            formState.errors.delete(fieldName);
            this.clearFieldError(formId, fieldName);
            this.showFieldSuccess(formId, fieldName);
        }

        // Update overall form validity
        this.updateFormValidity(formId);
        
        return errors.length === 0;
    }

    async executeValidationRule(rule, value) {
        const validator = this.validationRules.get(rule.type);
        
        if (!validator) {
            console.warn(`Unknown validation rule: ${rule.type}`);
            return true;
        }

        if (rule.type === 'custom' && rule.validator) {
            return rule.validator(value);
        }

        return validator(value, rule.value);
    }

    getErrorMessage(ruleType, ruleValue, customMessage) {
        if (customMessage) return customMessage;
        
        const messageTemplate = this.errorMessages.get(ruleType);
        
        if (typeof messageTemplate === 'function') {
            return messageTemplate(ruleValue);
        }
        
        return messageTemplate || `Validation failed for ${ruleType}`;
    }

    // UI Updates
    createErrorElement(field) {
        // Check if error element already exists
        const existingError = field.parentNode.querySelector('.field-error');
        if (existingError) return;

        const errorElement = document.createElement('div');
        errorElement.className = 'field-error';
        errorElement.style.display = 'none';
        
        // Insert after the field
        field.parentNode.insertBefore(errorElement, field.nextSibling);
    }

    showFieldErrors(formId, fieldName, errors) {
        const field = this.getField(formId, fieldName);
        if (!field) return;

        const errorElement = field.parentNode.querySelector('.field-error');
        if (errorElement) {
            errorElement.innerHTML = errors.map(error => `<span class="error-text">${error}</span>`).join('');
            errorElement.style.display = 'block';
        }

        // Add error class to field
        field.classList.add('field-invalid');
        field.classList.remove('field-valid', 'field-validating');
    }

    clearFieldError(formId, fieldName) {
        const field = this.getField(formId, fieldName);
        if (!field) return;

        const errorElement = field.parentNode.querySelector('.field-error');
        if (errorElement) {
            errorElement.style.display = 'none';
        }

        field.classList.remove('field-invalid');
    }

    showFieldSuccess(formId, fieldName) {
        const field = this.getField(formId, fieldName);
        if (!field) return;

        field.classList.add('field-valid');
        field.classList.remove('field-invalid', 'field-validating');
    }

    showFieldValidating(formId, fieldName) {
        const field = this.getField(formId, fieldName);
        if (!field) return;

        field.classList.add('field-validating');
        field.classList.remove('field-valid', 'field-invalid');
    }

    hideFieldValidating(formId, fieldName) {
        const field = this.getField(formId, fieldName);
        if (!field) return;

        field.classList.remove('field-validating');
    }

    // Form Validation
    async validateForm(formId) {
        const form = document.getElementById(formId);
        const formState = this.validationState.get(formId);
        
        if (!form || !formState) return false;

        formState.submitted = true;
        const fields = form.querySelectorAll('input, select, textarea');
        const validationPromises = [];

        // Validate all fields
        for (const field of fields) {
            const fieldName = field.name || field.id;
            if (fieldName) {
                this.markFieldAsTouched(formId, fieldName);
                validationPromises.push(
                    this.validateField(formId, fieldName, field.value)
                );
            }
        }

        const results = await Promise.all(validationPromises);
        const isValid = results.every(result => result);

        // TREC compliance validation for booking forms
        if (form.classList.contains('booking-form') || form.classList.contains('trec-form')) {
            const trecValid = await this.validateTRECCompliance(formId);
            return isValid && trecValid;
        }

        // Financial form validation
        if (form.classList.contains('financial-form')) {
            const financialValid = this.validateFinancialForm(formId);
            return isValid && financialValid;
        }

        return isValid;
    }

    async validateTRECCompliance(formId) {
        const form = document.getElementById(formId);
        const formState = this.validationState.get(formId);
        
        const errors = [];

        // Check required disclosures
        for (const disclosure of this.trecValidation.requiredDisclosures) {
            const checkbox = form.querySelector(`input[name="disclosure_${disclosure}"]`);
            if (!checkbox || !checkbox.checked) {
                errors.push(`${disclosure.replace('_', ' ')} disclosure must be acknowledged`);
            }
        }

        // Check required signatures
        for (const signature of this.trecValidation.requiredSignatures) {
            const signatureField = form.querySelector(`input[name="${signature}"]`);
            if (!signatureField || !signatureField.value.trim()) {
                errors.push(`${signature.replace('_', ' ')} is required`);
            }
        }

        // Validate important dates
        for (const [dateField, rules] of Object.entries(this.trecValidation.dateValidations)) {
            const dateInput = form.querySelector(`input[name="${dateField}"]`);
            if (dateInput && dateInput.value) {
                const dateValue = new Date(dateInput.value);
                
                if (rules.notPastDate && dateValue < new Date()) {
                    errors.push(`${dateField.replace('_', ' ')} cannot be in the past`);
                }
                
                if (rules.minDaysFromNow) {
                    const minDate = new Date();
                    minDate.setDate(minDate.getDate() + rules.minDaysFromNow);
                    if (dateValue < minDate) {
                        errors.push(`${dateField.replace('_', ' ')} must be at least ${rules.minDaysFromNow} days from now`);
                    }
                }
            } else if (rules.required) {
                errors.push(`${dateField.replace('_', ' ')} is required`);
            }
        }

        if (errors.length > 0) {
            formState.errors.set('trec_compliance', errors);
            this.showFormErrors(formId, errors);
            return false;
        }

        formState.errors.delete('trec_compliance');
        return true;
    }

    validateFinancialForm(formId) {
        const form = document.getElementById(formId);
        const formData = new FormData(form);
        const errors = [];

        // Validate debt-to-income ratio
        const income = parseFloat(formData.get('annualIncome')) || 0;
        const monthlyDebt = parseFloat(formData.get('monthlyDebt')) || 0;
        
        if (income > 0 && monthlyDebt > 0) {
            const monthlyIncome = income / 12;
            const dtiRatio = (monthlyDebt / monthlyIncome) * 100;
            
            if (dtiRatio > 43) { // Standard DTI limit
                errors.push('Debt-to-income ratio exceeds recommended limit (43%)');
            }
        }

        // Validate down payment vs loan amount
        const purchasePrice = parseFloat(formData.get('purchasePrice')) || 0;
        const downPayment = parseFloat(formData.get('downPayment')) || 0;
        
        if (purchasePrice > 0 && downPayment > 0) {
            const downPaymentRatio = (downPayment / purchasePrice) * 100;
            
            if (downPaymentRatio < 3) {
                errors.push('Down payment must be at least 3% of purchase price');
            }
            
            if (downPaymentRatio > 100) {
                errors.push('Down payment cannot exceed purchase price');
            }
        }

        // Credit score vs loan type validation
        const creditScore = parseInt(formData.get('creditScore')) || 0;
        const loanType = formData.get('loanType');
        
        if (creditScore > 0 && loanType) {
            const minScores = {
                'conventional': 620,
                'fha': 580,
                'va': 580,
                'usda': 640
            };
            
            if (minScores[loanType] && creditScore < minScores[loanType]) {
                errors.push(`Credit score too low for ${loanType.toUpperCase()} loan (minimum: ${minScores[loanType]})`);
            }
        }

        if (errors.length > 0) {
            const formState = this.validationState.get(formId);
            formState.errors.set('financial_validation', errors);
            this.showFormErrors(formId, errors);
            return false;
        }

        return true;
    }

    // Form Submission Handling
    async handleFormSubmission(event, formId) {
        event.preventDefault();
        
        const form = event.target;
        const submitButton = form.querySelector('button[type="submit"]');
        
        // Disable submit button
        if (submitButton) {
            submitButton.disabled = true;
            submitButton.dataset.originalText = submitButton.textContent;
            submitButton.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Validating...';
        }

        try {
            // Validate entire form
            const isValid = await this.validateForm(formId);
            
            if (!isValid) {
                this.showValidationSummary(formId);
                return false;
            }

            // Form is valid, proceed with submission
            await this.submitForm(formId, form);
            
        } catch (error) {
            console.error('Form submission error:', error);
            this.showSubmissionError(formId, error);
        } finally {
            // Re-enable submit button
            if (submitButton) {
                submitButton.disabled = false;
                submitButton.textContent = submitButton.dataset.originalText || 'Submit';
            }
        }
    }

    async submitForm(formId, form) {
        const formData = new FormData(form);
        const endpoint = form.action || '/api/forms/submit';
        const method = form.method || 'POST';

        try {
            const response = await fetch(endpoint, {
                method: method,
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: formData
            });

            if (!response.ok) {
                throw new Error(`Server error: ${response.status}`);
            }

            const result = await response.json();
            
            // Show success message
            this.showSubmissionSuccess(formId, result);
            
            // Track form submission
            this.trackFormSubmission(formId, 'success');
            
            // Handle redirect if specified
            if (result.redirectUrl) {
                setTimeout(() => {
                    window.location.href = result.redirectUrl;
                }, 2000);
            }

        } catch (error) {
            console.error('Form submission failed:', error);
            throw error;
        }
    }

    // Input Formatting
    formatCurrencyInput(input) {
        let value = input.value.replace(/[^\d.]/g, '');
        
        // Limit to 2 decimal places
        const parts = value.split('.');
        if (parts.length > 2) {
            value = parts[0] + '.' + parts[1];
        }
        if (parts[1] && parts[1].length > 2) {
            value = parts[0] + '.' + parts[1].substring(0, 2);
        }

        // Format with commas
        if (value) {
            const numberValue = parseFloat(value);
            if (!isNaN(numberValue)) {
                input.value = numberValue.toLocaleString('en-US', {
                    minimumFractionDigits: 0,
                    maximumFractionDigits: 2
                });
            }
        }
    }

    formatPhoneInput(input) {
        let value = input.value.replace(/\D/g, '');
        
        if (value.length >= 6) {
            value = value.substring(0, 3) + '-' + value.substring(3, 6) + '-' + value.substring(6, 10);
        } else if (value.length >= 3) {
            value = value.substring(0, 3) + '-' + value.substring(3);
        }
        
        input.value = value;
    }

    // Error Display
    showFormErrors(formId, errors) {
        const form = document.getElementById(formId);
        let errorContainer = form.querySelector('.form-errors');
        
        if (!errorContainer) {
            errorContainer = document.createElement('div');
            errorContainer.className = 'form-errors';
            form.insertBefore(errorContainer, form.firstChild);
        }

        errorContainer.innerHTML = `
            <div class="error-summary">
                <h4><i class="fas fa-exclamation-triangle"></i> Please correct the following errors:</h4>
                <ul>
                    ${errors.map(error => `<li>${error}</li>`).join('')}
                </ul>
            </div>
        `;

        errorContainer.style.display = 'block';
        
        // Scroll to top of form
        form.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }

    showValidationSummary(formId) {
        const formState = this.validationState.get(formId);
        const allErrors = [];
        
        for (const [fieldName, fieldErrors] of formState.errors) {
            allErrors.push(...fieldErrors);
        }

        if (allErrors.length > 0) {
            this.showFormErrors(formId, allErrors);
        }
    }

    showSubmissionSuccess(formId, result) {
        const form = document.getElementById(formId);
        let successContainer = form.querySelector('.form-success');
        
        if (!successContainer) {
            successContainer = document.createElement('div');
            successContainer.className = 'form-success';
            form.insertBefore(successContainer, form.firstChild);
        }

        successContainer.innerHTML = `
            <div class="success-message">
                <i class="fas fa-check-circle"></i>
                <h4>${result.message || 'Form submitted successfully!'}</h4>
                ${result.details ? `<p>${result.details}</p>` : ''}
            </div>
        `;

        successContainer.style.display = 'block';
        
        // Hide form fields
        const fieldGroups = form.querySelectorAll('.form-group, .form-row');
        fieldGroups.forEach(group => {
            group.style.display = 'none';
        });

        // Hide submit button
        const submitButton = form.querySelector('button[type="submit"]');
        if (submitButton) {
            submitButton.style.display = 'none';
        }
    }

    showSubmissionError(formId, error) {
        this.showFormErrors(formId, [`Submission failed: ${error.message}`]);
    }

    // Utility Methods
    getField(formId, fieldName) {
        const form = document.getElementById(formId);
        return form ? form.querySelector(`[name="${fieldName}"], #${fieldName}`) : null;
    }

    markFieldAsTouched(formId, fieldName) {
        const formState = this.validationState.get(formId);
        if (formState) {
            formState.touched.add(fieldName);
        }
    }

    updateFormValidity(formId) {
        const formState = this.validationState.get(formId);
        const form = document.getElementById(formId);
        
        if (!formState || !form) return;

        const hasErrors = formState.errors.size > 0;
        const hasAsyncValidating = formState.asyncValidating.size > 0;
        
        formState.isValid = !hasErrors && !hasAsyncValidating;
        
        // Update submit button state
        const submitButton = form.querySelector('button[type="submit"]');
        if (submitButton) {
            submitButton.disabled = !formState.isValid || hasAsyncValidating;
        }

        // Update form classes
        form.classList.toggle('form-valid', formState.isValid);
        form.classList.toggle('form-invalid', hasErrors);
        form.classList.toggle('form-validating', hasAsyncValidating);
    }

    generateFormId() {
        return 'form-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    trackFormSubmission(formId, status) {
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackEvent('form_submitted', {
                formId: formId,
                status: status,
                timestamp: new Date().toISOString()
            });
        }
    }

    setupEventListeners() {
        // Global form submission prevention for invalid forms
        document.addEventListener('submit', (e) => {
            const form = e.target;
            if (form.dataset.validate === 'true' || form.classList.contains('validate-form')) {
                const formId = form.id;
                const formState = this.validationState.get(formId);
                
                if (formState && !formState.isValid) {
                    e.preventDefault();
                    console.log('Form submission prevented due to validation errors');
                }
            }
        });

        // Dynamic form addition
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                mutation.addedNodes.forEach((node) => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        const forms = node.querySelectorAll ? 
                            node.querySelectorAll('form[data-validate="true"], .validate-form') : 
                            [];
                        
                        forms.forEach(form => {
                            this.setupFormValidation(form);
                        });
                    }
                });
            });
        });

        observer.observe(document.body, { childList: true, subtree: true });
    }

    // Public API Methods
    static validateForm(formId) {
        if (window.PropertyHubFormValidation) {
            return window.PropertyHubFormValidation.validateForm(formId);
        }
        return false;
    }

    static validateField(formId, fieldName, value) {
        if (window.PropertyHubFormValidation) {
            return window.PropertyHubFormValidation.validateField(formId, fieldName, value);
        }
        return false;
    }

    static clearFormErrors(formId) {
        if (window.PropertyHubFormValidation) {
            const formState = window.PropertyHubFormValidation.validationState.get(formId);
            if (formState) {
                formState.errors.clear();
                window.PropertyHubFormValidation.updateFormValidity(formId);
                
                // Clear all field errors
                const form = document.getElementById(formId);
                if (form) {
                    form.querySelectorAll('.field-error').forEach(errorEl => {
                        errorEl.style.display = 'none';
                    });
                    form.querySelectorAll('.field-invalid').forEach(field => {
                        field.classList.remove('field-invalid');
                    });
                }
            }
        }
    }

    static addCustomValidator(name, validatorFunction, errorMessage) {
        if (window.PropertyHubFormValidation) {
            window.PropertyHubFormValidation.validationRules.set(name, validatorFunction);
            window.PropertyHubFormValidation.errorMessages.set(name, errorMessage);
        }
    }

    static getFormState(formId) {
        if (window.PropertyHubFormValidation) {
            return window.PropertyHubFormValidation.validationState.get(formId);
        }
        return null;
    }

    // Advanced Validation Features
    validatePropertyForm(formId) {
        const form = document.getElementById(formId);
        const formData = new FormData(form);
        const errors = [];

        // Validate property pricing logic
        const listPrice = parseFloat(formData.get('listPrice'));
        const rentPrice = parseFloat(formData.get('rentPrice'));
        const squareFootage = parseFloat(formData.get('squareFootage'));

        if (listPrice && squareFootage) {
            const pricePerSqft = listPrice / squareFootage;
            if (pricePerSqft < 50 || pricePerSqft > 1000) {
                errors.push('Price per square foot seems unusual ($50-$1000 expected)');
            }
        }

        if (rentPrice && squareFootage) {
            const rentPerSqft = rentPrice / squareFootage;
            if (rentPerSqft < 0.5 || rentPerSqft > 10) {
                errors.push('Rent per square foot seems unusual ($0.50-$10.00 expected)');
            }
        }

        // Validate bedroom/bathroom logic
        const bedrooms = parseInt(formData.get('bedrooms'));
        const bathrooms = parseFloat(formData.get('bathrooms'));

        if (bedrooms && bathrooms && bathrooms > bedrooms + 1) {
            errors.push('Number of bathrooms typically should not exceed bedrooms + 1');
        }

        return errors;
    }

    validateBookingForm(formId) {
        const form = document.getElementById(formId);
        const formData = new FormData(form);
        const errors = [];

        // Validate showing date
        const showingDate = new Date(formData.get('showingDate'));
        const today = new Date();
        today.setHours(0, 0, 0, 0);

        if (showingDate && showingDate < today) {
            errors.push('Showing date must be in the future');
        }

        // Validate showing date is not too far in advance (6 months)
        const sixMonthsFromNow = new Date();
        sixMonthsFromNow.setMonth(sixMonthsFromNow.getMonth() + 6);
        if (showingDate && showingDate > sixMonthsFromNow) {
            errors.push('Showing date cannot be more than 6 months in the future');
        }

        // Validate showing time is provided
        const showingTime = formData.get('showingTime');
        if (!showingTime) {
            errors.push('Showing time is required');
        }

        // Validate attendee count (1-10 people for showings)
        const attendeeCount = parseInt(formData.get('attendeeCount'));
        if (attendeeCount && (attendeeCount < 1 || attendeeCount > 10)) {
            errors.push('Attendee count must be between 1 and 10');
        }

        // Validate duration (15-180 minutes)
        const duration = parseInt(formData.get('durationMinutes'));
        if (duration && (duration < 15 || duration > 180)) {
            errors.push('Duration must be between 15 and 180 minutes');
        }

        return errors;
    }

    // Real-time Field Enhancement
    setupFieldEnhancements() {
        // Auto-format currency fields
        document.querySelectorAll('.currency-input').forEach(input => {
            input.addEventListener('input', (e) => this.formatCurrencyInput(e.target));
        });

        // Auto-format phone fields
        document.querySelectorAll('.phone-input').forEach(input => {
            input.addEventListener('input', (e) => this.formatPhoneInput(e.target));
        });

        // Auto-capitalize address fields
        document.querySelectorAll('.address-input').forEach(input => {
            input.addEventListener('input', (e) => {
                const words = e.target.value.toLowerCase().split(' ');
                const capitalized = words.map(word => 
                    word.charAt(0).toUpperCase() + word.slice(1)
                );
                e.target.value = capitalized.join(' ');
            });
        });
    }

    // Advanced Validation Scenarios
    crossFieldValidation(formId) {
        const form = document.getElementById(formId);
        const formData = new FormData(form);
        const errors = [];

        // Example: Validate closing date based on contract date
        const contractDate = new Date(formData.get('contractDate'));
        const closingDate = new Date(formData.get('closingDate'));

        if (contractDate && closingDate) {
            const daysDiff = (closingDate - contractDate) / (1000 * 60 * 60 * 24);
            
            if (daysDiff < 7) {
                errors.push('Closing date must be at least 7 days after contract date');
            }
            
            if (daysDiff > 60) {
                errors.push('Closing date should typically be within 60 days of contract');
            }
        }

        return errors;
    }
}

// Initialize PropertyHub Form Validation when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.PropertyHubFormValidation = new PropertyHubFormValidation();
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubFormValidation;
}
