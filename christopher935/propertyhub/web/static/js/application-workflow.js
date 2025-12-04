// ApplicationWorkflow JavaScript for Christopher's specific business model

let draggedApplicant = null;
let currentAgentAssignment = null;

class ApplicationWorkflow {
    constructor() {
        this.properties = [];
        this.unassignedApplicants = [];
        this.initialize();
    }

    async initialize() {
        await this.loadData();
        this.setupDragAndDrop();
        this.setupEventListeners();
    }

    async loadData() {
        try {
            // Load properties with their application numbers
            const propertiesResponse = await fetch('/api/v1/application-workflow/properties');
            const propertiesData = await propertiesResponse.json();
            
            if (propertiesData.success) {
                this.properties = propertiesData.data.properties;
            }

            // Load unassigned applicants
            const applicantsResponse = await fetch('/api/v1/application-workflow/unassigned-applicants');
            const applicantsData = await applicantsResponse.json();
            
            if (applicantsData.success) {
                this.unassignedApplicants = applicantsData.data.unassigned_applicants;
            }

            this.renderInterface();
        } catch (error) {
            console.error('Failed to load application workflow data:', error);
            this.loadMockData();
        }
    }

    loadMockData() {
        // Mock data matching Christopher's business model
        this.unassignedApplicants = [
            {
                id: 1,
                applicant_name: "Stacy Jones",
                applicant_email: "stacy.jones@example.com",
                applicant_phone: "713-555-0123",
                property_address: "123 Main St",
                application_date: "2024-01-15T10:30:00Z",
                fub_match: true,
                source_email: "buildium_notification"
            },
            {
                id: 2,
                applicant_name: "Mike Wilson",
                applicant_email: "mike.wilson@example.com",
                applicant_phone: "281-555-0456",
                property_address: "456 Oak Ave",
                application_date: "2024-01-15T14:20:00Z",
                fub_match: false,
                source_email: "buildium_notification"
            },
            {
                id: 3,
                applicant_name: "Jennifer Lopez",
                applicant_email: "jennifer.l@example.com",
                applicant_phone: "713-555-0789",
                property_address: "123 Main St",
                application_date: "2024-01-15T16:45:00Z",
                fub_match: true,
                source_email: "buildium_notification"
            }
        ];

        this.properties = [
            {
                id: 1,
                property_address: "123 Main St",
                total_applications: 2,
                application_numbers: [
                    {
                        id: 1,
                        application_number: 1,
                        application_name: "Application 1",
                        status: "review",
                        assigned_agent_name: "John Smith",
                        assigned_agent_phone: "713-555-0001",
                        assigned_agent_email: "john@example.com",
                        applicant_count: 2,
                        applicants: [
                            {
                                id: 4,
                                applicant_name: "Sarah Davis",
                                applicant_email: "sarah@example.com",
                                fub_match: true
                            },
                            {
                                id: 5,
                                applicant_name: "Tom Brown",
                                applicant_email: "tom@example.com", 
                                fub_match: false
                            }
                        ]
                    },
                    {
                        id: 2,
                        application_number: 2,
                        application_name: "Application 2", 
                        status: "submitted",
                        assigned_agent_name: "",
                        assigned_agent_phone: "",
                        assigned_agent_email: "",
                        applicant_count: 1,
                        applicants: [
                            {
                                id: 6,
                                applicant_name: "Lisa Wilson",
                                applicant_email: "lisa@example.com",
                                fub_match: true
                            }
                        ]
                    }
                ]
            },
            {
                id: 2,
                property_address: "456 Oak Ave",
                total_applications: 1,
                application_numbers: [
                    {
                        id: 3,
                        application_number: 1,
                        application_name: "Application 1",
                        status: "further_review",
                        assigned_agent_name: "Mary Johnson",
                        assigned_agent_phone: "281-555-0002",
                        assigned_agent_email: "mary@example.com",
                        applicant_count: 1,
                        applicants: [
                            {
                                id: 7,
                                applicant_name: "David Kim",
                                applicant_email: "david@example.com",
                                fub_match: true
                            }
                        ]
                    }
                ]
            }
        ];

        this.renderInterface();
    }

    renderInterface() {
        this.renderUnassignedApplicants();
        this.renderProperties();
    }

    renderUnassignedApplicants() {
        const container = document.getElementById('unassignedApplicants');
        document.getElementById('unassignedCount').textContent = this.unassignedApplicants.length;
        
        container.innerHTML = this.unassignedApplicants.map(applicant => `
            <div class="applicant-card" draggable="true" data-applicant-id="${applicant.id}">
                <div class="applicant-name">${applicant.applicant_name}</div>
                <div class="applicant-contact">
                    📧 ${applicant.applicant_email}<br>
                    📞 ${applicant.applicant_phone}
                </div>
                <div style="font-size: 11px; color: #64748b; margin: 4px 0;">
                    🏠 ${applicant.property_address}
                </div>
                <div class="applicant-meta">
                    <span class="fub-badge ${applicant.fub_match ? 'fub-matched' : 'fub-unmatched'}">
                        ${applicant.fub_match ? '✅ FUB' : '❌ No FUB'}
                    </span>
                    <span class="applicant-date">${this.formatDate(applicant.application_date)}</span>
                </div>
            </div>
        `).join('');
    }

    renderProperties() {
        const container = document.getElementById('propertiesContainer');
        
        container.innerHTML = this.properties.map(property => `
            <div class="property-card">
                <div class="property-header">
                    <div class="property-title">🏠 ${property.property_address}</div>
                    <div class="property-meta">
                        ${property.application_numbers ? property.application_numbers.length : 0} applications • 
                        ${this.getTotalApplicants(property)} total applicants
                    </div>
                </div>
                
                <div class="applications-grid">
                    ${this.renderApplicationNumbers(property)}
                    <div class="add-application-btn" onclick="addApplicationNumber(${property.id}, '${property.property_address}')">
                        <div class="add-application-icon">➕</div>
                        <div>Add Application ${(property.application_numbers || []).length + 1}</div>
                    </div>
                </div>
            </div>
        `).join('');
    }

    renderApplicationNumbers(property) {
        if (!property.application_numbers || property.application_numbers.length === 0) {
            return '';
        }

        return property.application_numbers.map(app => `
            <div class="application-column" data-application-id="${app.id}">
                <div class="application-header">
                    <div class="application-title">
                        <span>${app.application_name}</span>
                        <span class="application-status status-${app.status.replace('_', '-')}">${app.status.replace('_', ' ')}</span>
                    </div>
                    
                    <select class="status-dropdown" onchange="updateApplicationStatus(${app.id}, this.value)">
                        <option value="submitted" ${app.status === 'submitted' ? 'selected' : ''}>Submitted</option>
                        <option value="review" ${app.status === 'review' ? 'selected' : ''}>Review</option>
                        <option value="further_review" ${app.status === 'further_review' ? 'selected' : ''}>Further Review</option>
                        <option value="rental_history_received" ${app.status === 'rental_history_received' ? 'selected' : ''}>Rental History Received</option>
                        <option value="approved" ${app.status === 'approved' ? 'selected' : ''}>Approved</option>
                        <option value="denied" ${app.status === 'denied' ? 'selected' : ''}>Denied</option>
                        <option value="backup" ${app.status === 'backup' ? 'selected' : ''}>Backup</option>
                        <option value="cancelled" ${app.status === 'cancelled' ? 'selected' : ''}>Cancelled</option>
                    </select>
                    
                    <div class="agent-assignment">
                        ${app.assigned_agent_name ? `
                            <div class="agent-info">
                                👤 ${app.assigned_agent_name}<br>
                                📞 ${app.assigned_agent_phone}
                            </div>
                            <button class="assign-agent-btn" onclick="changeAgent(${app.id})">Change Agent</button>
                        ` : `
                            <div class="no-agent">No agent assigned</div>
                            <button class="assign-agent-btn" onclick="assignAgent(${app.id})">Assign Agent</button>
                        `}
                    </div>
                </div>
                
                <div class="application-content">
                    ${app.applicants && app.applicants.length > 0 ? 
                        app.applicants.map(applicant => `
                            <div class="applicant-item" draggable="true" data-applicant-id="${applicant.id}">
                                <div class="applicant-item-name">${applicant.applicant_name}</div>
                                <div class="applicant-item-contact">${applicant.applicant_email}</div>
                                ${applicant.fub_match ? '<span class="fub-badge fub-matched">✅ FUB</span>' : ''}
                            </div>
                        `).join('') : 
                        '<div class="empty-application">Drop applicants here</div>'
                    }
                </div>
            </div>
        `).join('');
    }

    getTotalApplicants(property) {
        if (!property.application_numbers) return 0;
        return property.application_numbers.reduce((total, app) => total + (app.applicant_count || 0), 0);
    }

    formatDate(dateStr) {
        return new Date(dateStr).toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    setupDragAndDrop() {
        // Set up drag and drop for applicants
        document.addEventListener('dragstart', (e) => {
            if (e.target.classList.contains('applicant-card') || e.target.classList.contains('applicant-item')) {
                draggedApplicant = {
                    id: e.target.dataset.applicantId,
                    element: e.target,
                    sourceType: e.target.classList.contains('applicant-card') ? 'unassigned' : 'assigned'
                };
                e.target.classList.add('dragging');
            }
        });

        document.addEventListener('dragend', (e) => {
            if (e.target.classList.contains('applicant-card') || e.target.classList.contains('applicant-item')) {
                e.target.classList.remove('dragging');
                draggedApplicant = null;
            }
        });

        document.addEventListener('dragover', (e) => {
            e.preventDefault();
        });

        document.addEventListener('dragenter', (e) => {
            if (e.target.closest('.application-column')) {
                e.target.closest('.application-column').classList.add('drag-over');
            }
        });

        document.addEventListener('dragleave', (e) => {
            if (e.target.closest('.application-column')) {
                e.target.closest('.application-column').classList.remove('drag-over');
            }
        });

        document.addEventListener('drop', (e) => {
            e.preventDefault();
            
            const applicationColumn = e.target.closest('.application-column');
            if (applicationColumn && draggedApplicant) {
                applicationColumn.classList.remove('drag-over');
                const applicationId = applicationColumn.dataset.applicationId;
                this.moveApplicantToApplication(draggedApplicant.id, applicationId);
            }
        });
    }

    setupEventListeners() {
        // Agent modal form submission
        document.getElementById('agentForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.submitAgentAssignment();
        });
    }

    async moveApplicantToApplication(applicantId, applicationId) {
        try {
            const response = await fetch('/api/v1/application-workflow/move-applicant', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    applicant_id: parseInt(applicantId),
                    target_application_id: parseInt(applicationId),
                    moved_by: 'Admin',
                    reason: 'Drag and drop assignment'
                })
            });

            const result = await response.json();
            
            if (result.success) {
                this.showMessage('✅ ' + result.message, 'success');
                await this.loadData(); // Refresh data
            } else {
                this.showMessage('❌ Failed to move applicant', 'error');
            }
        } catch (error) {
            console.error('Failed to move applicant:', error);
            this.showMessage('❌ Network error', 'error');
        }
    }

    async submitAgentAssignment() {
        const agentName = document.getElementById('agentName').value;
        const agentPhone = document.getElementById('agentPhone').value;
        const agentEmail = document.getElementById('agentEmail').value;

        try {
            const response = await fetch('/api/v1/application-workflow/assign-agent', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    application_number_id: currentAgentAssignment,
                    agent_name: agentName,
                    agent_phone: agentPhone,
                    agent_email: agentEmail,
                    assigned_by: 'Admin'
                })
            });

            const result = await response.json();
            
            if (result.success) {
                this.showMessage('✅ ' + result.message, 'success');
                closeAgentModal();
                await this.loadData();
            } else {
                this.showMessage('❌ Failed to assign agent', 'error');
            }
        } catch (error) {
            console.error('Failed to assign agent:', error);
            this.showMessage('❌ Network error', 'error');
        }
    }

    showMessage(message, type) {
        // Simple toast notification
        const toast = document.createElement('div');
        toast.className = `fixed top-4 right-4 px-4 py-2 rounded-md text-white z-50 ${type === 'success' ? 'bg-green-600' : 'bg-red-600'}`;
        toast.textContent = message;
        toast.style.cssText = 'position: fixed; top: 1rem; right: 1rem; padding: 0.5rem 1rem; border-radius: 0.375rem; color: white; z-index: 50;';
        document.body.appendChild(toast);
        
        setTimeout(() => {
            if (document.body.contains(toast)) {
                document.body.removeChild(toast);
            }
        }, 3000);
    }
}

// Global functions for template interactions

async function addApplicationNumber(propertyId, propertyAddress) {
    try {
        const response = await fetch(`/api/v1/application-workflow/properties/${propertyId}/applications`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                property_address: propertyAddress
            })
        });

        const result = await response.json();
        
        if (result.success) {
            window.workflow.showMessage('✅ ' + result.data.message, 'success');
            await window.workflow.loadData();
        } else {
            window.workflow.showMessage('❌ Failed to add application', 'error');
        }
    } catch (error) {
        console.error('Failed to add application:', error);
        window.workflow.showMessage('❌ Network error', 'error');
    }
}

function assignAgent(applicationId) {
    currentAgentAssignment = applicationId;
    document.getElementById('agentModal').style.display = 'block';
}

function changeAgent(applicationId) {
    currentAgentAssignment = applicationId;
    // Pre-fill form with existing agent data if available
    document.getElementById('agentModal').style.display = 'block';
}

function closeAgentModal() {
    document.getElementById('agentModal').style.display = 'none';
    currentAgentAssignment = null;
    document.getElementById('agentForm').reset();
}

async function updateApplicationStatus(applicationId, newStatus) {
    try {
        const response = await fetch('/api/v1/application-workflow/update-status', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                application_number_id: applicationId,
                status: newStatus,
                updated_by: 'Admin',
                reason: 'Manual status update'
            })
        });

        const result = await response.json();
        
        if (result.success) {
            window.workflow.showMessage('✅ Status updated', 'success');
            await window.workflow.loadData();
        } else {
            window.workflow.showMessage('❌ Failed to update status', 'error');
        }
    } catch (error) {
        console.error('Failed to update status:', error);
        window.workflow.showMessage('❌ Network error', 'error');
    }
}

async function refreshUnassigned() {
    await window.workflow.loadData();
    window.workflow.showMessage('🔄 Refreshed unassigned applicants', 'success');
}

async function refreshData() {
    await window.workflow.loadData();
    window.workflow.showMessage('🔄 Data refreshed', 'success');
}

function showAgentManager() {
    // Placeholder for agent management interface
    alert('Agent management interface - coming soon');
}

function exportApplications() {
    // Placeholder for export functionality
    alert('Export functionality - coming soon');
}

// Initialize the application workflow system
document.addEventListener('DOMContentLoaded', function() {
    window.workflow = new ApplicationWorkflow();
});