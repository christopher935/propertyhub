/**
 * PropertyHub File Upload System
 * Multi-file upload, image processing, document management
 * Real estate document handling with TREC compliance and security
 */

class PropertyHubFileUpload {
    constructor() {
        this.uploadQueue = [];
        this.activeUploads = new Map();
        this.maxConcurrentUploads = 3;
        this.maxFileSize = 50 * 1024 * 1024; // 50MB
        this.maxTotalSize = 500 * 1024 * 1024; // 500MB
        this.supportedImageTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/gif'];
        this.supportedDocumentTypes = [
            'application/pdf',
            'application/msword',
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
            'application/vnd.ms-excel',
            'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
            'text/plain'
        ];
        
        // File categories for real estate
        this.fileCategories = {
            PROPERTY_PHOTOS: 'property_photos',
            FLOOR_PLANS: 'floor_plans',
            DOCUMENTS: 'documents',
            TREC_FORMS: 'trec_forms',
            INSPECTION_REPORTS: 'inspection_reports',
            APPRAISAL_REPORTS: 'appraisal_reports',
            CONTRACTS: 'contracts',
            DISCLOSURES: 'disclosures',
            FINANCIAL_DOCUMENTS: 'financial_documents',
            INSURANCE_DOCUMENTS: 'insurance_documents'
        };

        // Image processing settings
        this.imageSettings = {
            thumbnail: { width: 300, height: 200, quality: 0.8 },
            preview: { width: 800, height: 600, quality: 0.9 },
            fullsize: { width: 1920, height: 1440, quality: 0.95 }
        };

        this.uploadZones = new Map();
        this.fileValidators = new Map();
        
        this.init();
    }

    async init() {
        console.log('📁 PropertyHub File Upload initializing...');
        
        this.setupFileValidators();
        this.setupUploadZones();
        this.setupEventListeners();
        this.initializeImageProcessing();
        
        console.log('✅ PropertyHub File Upload ready');
    }

    // Upload Zone Management
    setupUploadZones() {
        const uploadZones = document.querySelectorAll('.file-upload-zone');
        
        uploadZones.forEach(zone => {
            this.initializeUploadZone(zone);
        });
    }

    initializeUploadZone(zoneElement) {
        const zoneId = zoneElement.id || this.generateZoneId();
        const category = zoneElement.dataset.category || this.fileCategories.DOCUMENTS;
        const multiple = zoneElement.dataset.multiple !== 'false';
        const accept = zoneElement.dataset.accept || '*/*';

        const zoneConfig = {
            element: zoneElement,
            category: category,
            multiple: multiple,
            accept: accept,
            files: [],
            maxFiles: parseInt(zoneElement.dataset.maxFiles) || (multiple ? 50 : 1)
        };

        this.uploadZones.set(zoneId, zoneConfig);
        
        // Render upload zone UI
        this.renderUploadZone(zoneId, zoneConfig);
        
        // Setup drag and drop
        this.setupDragAndDrop(zoneElement, zoneId);
    }

    renderUploadZone(zoneId, config) {
        const zoneElement = config.element;
        
        zoneElement.innerHTML = `
            <div class="upload-zone-container">
                <div class="upload-dropzone" id="dropzone-${zoneId}">
                    <div class="upload-icon">
                        <i class="fas fa-cloud-upload-alt fa-3x"></i>
                    </div>
                    <div class="upload-text">
                        <h4>Drop files here or click to browse</h4>
                        <p>
                            ${this.getAcceptedTypesText(config.category)} 
                            ${config.multiple ? `(up to ${config.maxFiles} files)` : ''}
                        </p>
                        <p class="file-size-limit">Maximum file size: ${this.formatFileSize(this.maxFileSize)}</p>
                    </div>
                    <input type="file" 
                           id="file-input-${zoneId}" 
                           class="file-input-hidden"
                           ${config.multiple ? 'multiple' : ''}
                           accept="${config.accept}">
                </div>
                
                <div class="upload-progress-container" id="progress-${zoneId}" style="display: none;">
                    <div class="upload-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: 0%"></div>
                        </div>
                        <div class="progress-text">
                            <span class="progress-status">Preparing upload...</span>
                            <span class="progress-percentage">0%</span>
                        </div>
                    </div>
                </div>
                
                <div class="uploaded-files-container" id="files-${zoneId}">
                    <!-- Uploaded files will appear here -->
                </div>
            </div>
        `;

        // Add click handler for browse button
        const dropzone = zoneElement.querySelector(`#dropzone-${zoneId}`);
        const fileInput = zoneElement.querySelector(`#file-input-${zoneId}`);
        
        dropzone.addEventListener('click', () => {
            fileInput.click();
        });

        fileInput.addEventListener('change', (e) => {
            this.handleFileSelection(zoneId, e.target.files);
        });
    }

    setupDragAndDrop(zoneElement, zoneId) {
        const dropzone = zoneElement.querySelector(`#dropzone-${zoneId}`);
        
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            dropzone.addEventListener(eventName, this.preventDefaults, false);
        });

        ['dragenter', 'dragover'].forEach(eventName => {
            dropzone.addEventListener(eventName, () => {
                dropzone.classList.add('drag-over');
            }, false);
        });

        ['dragleave', 'drop'].forEach(eventName => {
            dropzone.addEventListener(eventName, () => {
                dropzone.classList.remove('drag-over');
            }, false);
        });

        dropzone.addEventListener('drop', (e) => {
            const files = e.dataTransfer.files;
            this.handleFileSelection(zoneId, files);
        }, false);
    }

    preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    // File Validation
    setupFileValidators() {
        // Property photos validator
        this.fileValidators.set(this.fileCategories.PROPERTY_PHOTOS, {
            types: this.supportedImageTypes,
            maxSize: 10 * 1024 * 1024, // 10MB
            minDimensions: { width: 800, height: 600 },
            requiresProcessing: true
        });

        // Documents validator
        this.fileValidators.set(this.fileCategories.DOCUMENTS, {
            types: [...this.supportedImageTypes, ...this.supportedDocumentTypes],
            maxSize: this.maxFileSize,
            requiresProcessing: false
        });

        // TREC forms validator
        this.fileValidators.set(this.fileCategories.TREC_FORMS, {
            types: ['application/pdf', ...this.supportedDocumentTypes],
            maxSize: this.maxFileSize,
            requiresCompliance: true,
            requiresProcessing: false
        });

        // Financial documents validator
        this.fileValidators.set(this.fileCategories.FINANCIAL_DOCUMENTS, {
            types: this.supportedDocumentTypes,
            maxSize: this.maxFileSize,
            requiresEncryption: true,
            requiresProcessing: false
        });
    }

    validateFile(file, category) {
        const validator = this.fileValidators.get(category) || this.fileValidators.get(this.fileCategories.DOCUMENTS);
        const errors = [];

        // Check file type
        if (validator.types && !validator.types.includes(file.type)) {
            errors.push(`File type ${file.type} is not supported for ${category}`);
        }

        // Check file size
        const maxSize = validator.maxSize || this.maxFileSize;
        if (file.size > maxSize) {
            errors.push(`File size exceeds maximum of ${this.formatFileSize(maxSize)}`);
        }

        // Check for malicious files
        if (this.isSuspiciousFile(file)) {
            errors.push('File appears to contain suspicious content');
        }

        return {
            valid: errors.length === 0,
            errors: errors,
            validator: validator
        };
    }

    isSuspiciousFile(file) {
        // Basic security checks
        const suspiciousExtensions = ['.exe', '.bat', '.cmd', '.scr', '.vbs', '.js'];
        const filename = file.name.toLowerCase();
        
        return suspiciousExtensions.some(ext => filename.endsWith(ext));
    }

    // File Processing
    async handleFileSelection(zoneId, files) {
        const config = this.uploadZones.get(zoneId);
        if (!config) return;

        const fileArray = Array.from(files);
        
        // Validate total number of files
        if (!config.multiple && fileArray.length > 1) {
            this.showError('Only one file is allowed in this area');
            return;
        }

        if (config.files.length + fileArray.length > config.maxFiles) {
            this.showError(`Maximum of ${config.maxFiles} files allowed`);
            return;
        }

        // Validate each file
        const validFiles = [];
        const invalidFiles = [];

        for (const file of fileArray) {
            const validation = this.validateFile(file, config.category);
            
            if (validation.valid) {
                validFiles.push({
                    file: file,
                    id: this.generateFileId(),
                    category: config.category,
                    validator: validation.validator,
                    status: 'pending'
                });
            } else {
                invalidFiles.push({
                    file: file,
                    errors: validation.errors
                });
            }
        }

        // Show validation errors
        if (invalidFiles.length > 0) {
            this.showValidationErrors(invalidFiles);
        }

        // Process valid files
        if (validFiles.length > 0) {
            await this.processFiles(zoneId, validFiles);
        }
    }

    async processFiles(zoneId, fileObjects) {
        for (const fileObj of fileObjects) {
            // Add to zone's file list
            const config = this.uploadZones.get(zoneId);
            config.files.push(fileObj);

            // Show file in UI
            this.addFileToUI(zoneId, fileObj);

            // Process file if needed
            if (fileObj.validator.requiresProcessing) {
                await this.processImageFile(fileObj);
            }

            // Add to upload queue
            this.addToUploadQueue(zoneId, fileObj);
        }

        // Start uploading
        this.processUploadQueue();
    }

    async processImageFile(fileObj) {
        if (!this.supportedImageTypes.includes(fileObj.file.type)) {
            return;
        }

        try {
            fileObj.status = 'processing';
            this.updateFileStatus(fileObj.id, 'Processing image...');

            // Create different sizes
            const sizes = ['thumbnail', 'preview', 'fullsize'];
            fileObj.processedImages = {};

            for (const size of sizes) {
                const settings = this.imageSettings[size];
                const processedBlob = await this.resizeImage(fileObj.file, settings);
                fileObj.processedImages[size] = processedBlob;
            }

            // Generate EXIF data if available
            fileObj.metadata = await this.extractImageMetadata(fileObj.file);

            fileObj.status = 'processed';
            this.updateFileStatus(fileObj.id, 'Ready to upload');

        } catch (error) {
            console.error('Error processing image:', error);
            fileObj.status = 'error';
            this.updateFileStatus(fileObj.id, 'Processing failed');
        }
    }

    async resizeImage(file, settings) {
        return new Promise((resolve, reject) => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            const img = new Image();

            img.onload = () => {
                // Calculate new dimensions maintaining aspect ratio
                const { width: targetWidth, height: targetHeight, quality } = settings;
                const aspectRatio = img.width / img.height;

                let newWidth = targetWidth;
                let newHeight = targetHeight;

                if (aspectRatio > targetWidth / targetHeight) {
                    newHeight = targetWidth / aspectRatio;
                } else {
                    newWidth = targetHeight * aspectRatio;
                }

                canvas.width = newWidth;
                canvas.height = newHeight;

                // Draw resized image
                ctx.drawImage(img, 0, 0, newWidth, newHeight);

                // Convert to blob
                canvas.toBlob(resolve, file.type, quality);
            };

            img.onerror = reject;
            img.src = URL.createObjectURL(file);
        });
    }

    async extractImageMetadata(file) {
        // Basic metadata extraction (would use a library like exif-js in production)
        const metadata = {
            fileName: file.name,
            fileSize: file.size,
            fileType: file.type,
            lastModified: new Date(file.lastModified).toISOString()
        };

        // Try to get image dimensions
        try {
            const dimensions = await this.getImageDimensions(file);
            metadata.width = dimensions.width;
            metadata.height = dimensions.height;
        } catch (error) {
            console.error('Error getting image dimensions:', error);
        }

        return metadata;
    }

    async getImageDimensions(file) {
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.onload = () => {
                resolve({ width: img.width, height: img.height });
                URL.revokeObjectURL(img.src);
            };
            img.onerror = reject;
            img.src = URL.createObjectURL(file);
        });
    }

    // Upload Management
    addToUploadQueue(zoneId, fileObj) {
        this.uploadQueue.push({
            zoneId: zoneId,
            fileObj: fileObj,
            retries: 0,
            maxRetries: 3
        });
    }

    async processUploadQueue() {
        while (this.uploadQueue.length > 0 && this.activeUploads.size < this.maxConcurrentUploads) {
            const uploadTask = this.uploadQueue.shift();
            this.startUpload(uploadTask);
        }
    }

    async startUpload(uploadTask) {
        const { zoneId, fileObj } = uploadTask;
        const uploadId = this.generateUploadId();
        
        this.activeUploads.set(uploadId, uploadTask);
        
        try {
            fileObj.status = 'uploading';
            this.updateFileStatus(fileObj.id, 'Uploading...');
            
            const result = await this.uploadFile(fileObj, (progress) => {
                this.updateUploadProgress(fileObj.id, progress);
            });

            fileObj.status = 'completed';
            fileObj.uploadResult = result;
            this.updateFileStatus(fileObj.id, 'Upload complete');
            
            // Track successful upload
            this.trackFileUpload(fileObj, 'success');
            
        } catch (error) {
            console.error('Upload failed:', error);
            
            if (uploadTask.retries < uploadTask.maxRetries) {
                uploadTask.retries++;
                this.uploadQueue.unshift(uploadTask); // Retry
                fileObj.status = 'retrying';
                this.updateFileStatus(fileObj.id, `Retrying... (${uploadTask.retries}/${uploadTask.maxRetries})`);
            } else {
                fileObj.status = 'error';
                fileObj.error = error.message;
                this.updateFileStatus(fileObj.id, 'Upload failed');
                this.trackFileUpload(fileObj, 'error', error);
            }
        } finally {
            this.activeUploads.delete(uploadId);
            this.processUploadQueue(); // Process next in queue
        }
    }

    async uploadFile(fileObj, progressCallback) {
        const formData = new FormData();
        
        // Add original file
        formData.append('file', fileObj.file);
        formData.append('category', fileObj.category);
        formData.append('metadata', JSON.stringify(fileObj.metadata || {}));
        
        // Add processed images if available
        if (fileObj.processedImages) {
            for (const [size, blob] of Object.entries(fileObj.processedImages)) {
                formData.append(`processed_${size}`, blob, `${size}_${fileObj.file.name}`);
            }
        }

        // Add TREC compliance info if needed
        if (fileObj.validator.requiresCompliance) {
            formData.append('trec_compliance', 'true');
        }

        // Add encryption flag if needed
        if (fileObj.validator.requiresEncryption) {
            formData.append('encrypt', 'true');
        }

        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const progress = (e.loaded / e.total) * 100;
                    progressCallback(progress);
                }
            });

            xhr.addEventListener('load', () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    try {
                        const result = JSON.parse(xhr.responseText);
                        resolve(result);
                    } catch (error) {
                        reject(new Error('Invalid server response'));
                    }
                } else {
                    reject(new Error(`Upload failed with status ${xhr.status}`));
                }
            });

            xhr.addEventListener('error', () => {
                reject(new Error('Network error during upload'));
            });

            xhr.open('POST', '/api/files/upload');
            
            // Add auth header
            if (window.PropertyHubAuth && window.PropertyHubAuth.token) {
                xhr.setRequestHeader('Authorization', `Bearer ${window.PropertyHubAuth.token}`);
            }
            
            xhr.send(formData);
        });
    }

    // UI Management
    addFileToUI(zoneId, fileObj) {
        const filesContainer = document.getElementById(`files-${zoneId}`);
        if (!filesContainer) return;

        const fileElement = document.createElement('div');
        fileElement.className = 'uploaded-file-item';
        fileElement.id = `file-${fileObj.id}`;
        
        fileElement.innerHTML = `
            <div class="file-preview">
                ${this.isImageFile(fileObj.file) ? 
                    `<img src="${URL.createObjectURL(fileObj.file)}" alt="Preview" class="file-thumbnail">` :
                    `<div class="file-icon"><i class="${this.getFileIcon(fileObj.file.type)}"></i></div>`
                }
            </div>
            
            <div class="file-info">
                <div class="file-name" title="${fileObj.file.name}">${fileObj.file.name}</div>
                <div class="file-details">
                    <span class="file-size">${this.formatFileSize(fileObj.file.size)}</span>
                    <span class="file-type">${fileObj.file.type.split('/')[1]?.toUpperCase() || 'File'}</span>
                </div>
                <div class="file-status" id="status-${fileObj.id}">Pending</div>
            </div>
            
            <div class="file-progress" id="progress-${fileObj.id}">
                <div class="progress-bar">
                    <div class="progress-fill" style="width: 0%"></div>
                </div>
                <span class="progress-text">0%</span>
            </div>
            
            <div class="file-actions">
                <button class="btn btn-sm btn-outline file-remove" onclick="PropertyHubFileUpload.removeFile('${zoneId}', '${fileObj.id}')">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;

        filesContainer.appendChild(fileElement);

        // Update dropzone visibility
        this.updateDropzoneVisibility(zoneId);
    }

    updateFileStatus(fileId, status) {
        const statusElement = document.getElementById(`status-${fileId}`);
        if (statusElement) {
            statusElement.textContent = status;
        }
    }

    updateUploadProgress(fileId, progress) {
        const progressContainer = document.getElementById(`progress-${fileId}`);
        if (!progressContainer) return;

        const progressFill = progressContainer.querySelector('.progress-fill');
        const progressText = progressContainer.querySelector('.progress-text');

        if (progressFill) {
            progressFill.style.width = `${progress}%`;
        }
        
        if (progressText) {
            progressText.textContent = `${Math.round(progress)}%`;
        }
    }

    updateDropzoneVisibility(zoneId) {
        const config = this.uploadZones.get(zoneId);
        const dropzone = document.getElementById(`dropzone-${zoneId}`);
        
        if (dropzone && config) {
            // Hide dropzone if max files reached and not multiple
            if (!config.multiple && config.files.length >= 1) {
                dropzone.style.display = 'none';
            } else if (config.files.length >= config.maxFiles) {
                dropzone.style.display = 'none';
            } else {
                dropzone.style.display = 'block';
            }
        }
    }

    // File Management
    removeFile(zoneId, fileId) {
        const config = this.uploadZones.get(zoneId);
        if (!config) return;

        // Remove from config
        config.files = config.files.filter(f => f.id !== fileId);

        // Remove from UI
        const fileElement = document.getElementById(`file-${fileId}`);
        if (fileElement) {
            fileElement.remove();
        }

        // Update dropzone visibility
        this.updateDropzoneVisibility(zoneId);

        // Cancel upload if in progress
        this.cancelUpload(fileId);
    }

    cancelUpload(fileId) {
        // Find and remove from upload queue
        this.uploadQueue = this.uploadQueue.filter(task => task.fileObj.id !== fileId);

        // Find and cancel active upload
        for (const [uploadId, uploadTask] of this.activeUploads) {
            if (uploadTask.fileObj.id === fileId) {
                // Cancel the upload (this would need xhr reference in real implementation)
                this.activeUploads.delete(uploadId);
                break;
            }
        }
    }

    // Document Management
    async organizeDocuments(propertyId, documents) {
        // Organize uploaded documents by category
        const organizedDocs = {};
        
        for (const doc of documents) {
            const category = doc.category || this.fileCategories.DOCUMENTS;
            if (!organizedDocs[category]) {
                organizedDocs[category] = [];
            }
            organizedDocs[category].push(doc);
        }

        // Update document index
        try {
            await fetch(`/api/properties/${propertyId}/documents/organize`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({ documents: organizedDocs })
            });
        } catch (error) {
            console.error('Error organizing documents:', error);
        }
    }

    async generateTRECDocuments(propertyId, documentTypes) {
        // Generate required TREC forms
        try {
            const response = await fetch('/api/trec/generate-documents', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    propertyId: propertyId,
                    documentTypes: documentTypes
                })
            });

            if (response.ok) {
                const result = await response.json();
                return result.documents;
            }
            
            throw new Error('Failed to generate TREC documents');
        } catch (error) {
            console.error('Error generating TREC documents:', error);
            throw error;
        }
    }

    // Event Listeners
    setupEventListeners() {
        // Global file drop prevention
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            document.addEventListener(eventName, (e) => {
                // Only prevent default if not in an upload zone
                if (!e.target.closest('.file-upload-zone')) {
                    e.preventDefault();
                    e.stopPropagation();
                }
            }, false);
        });

        // Paste handling for image uploads
        document.addEventListener('paste', (e) => {
            const activeZone = document.querySelector('.file-upload-zone:focus-within');
            if (activeZone && e.clipboardData && e.clipboardData.files.length > 0) {
                const zoneId = activeZone.id;
                this.handleFileSelection(zoneId, e.clipboardData.files);
            }
        });
    }

    // Utility Methods
    generateZoneId() {
        return 'upload-zone-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    generateFileId() {
        return 'file-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    generateUploadId() {
        return 'upload-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    isImageFile(file) {
        return this.supportedImageTypes.includes(file.type);
    }

    getFileIcon(mimeType) {
        const iconMap = {
            'application/pdf': 'fas fa-file-pdf',
            'application/msword': 'fas fa-file-word',
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'fas fa-file-word',
            'application/vnd.ms-excel': 'fas fa-file-excel',
            'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'fas fa-file-excel',
            'text/plain': 'fas fa-file-alt',
            'image/jpeg': 'fas fa-file-image',
            'image/png': 'fas fa-file-image',
            'image/gif': 'fas fa-file-image'
        };

        return iconMap[mimeType] || 'fas fa-file';
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    getAcceptedTypesText(category) {
        switch (category) {
            case this.fileCategories.PROPERTY_PHOTOS:
                return 'Images (JPG, PNG, WebP, GIF)';
            case this.fileCategories.TREC_FORMS:
                return 'PDF and Office documents';
            case this.fileCategories.FINANCIAL_DOCUMENTS:
                return 'PDF and Office documents';
            default:
                return 'Images and documents';
        }
    }

    showError(message) {
        if (window.PropertyHubNotifications) {
            window.PropertyHubNotifications.showToastNotification({
                title: 'Upload Error',
                message: message,
                type: 'error',
                duration: 5000
            });
        } else {
            alert(message);
        }
    }

    showValidationErrors(invalidFiles) {
        const errorMessages = invalidFiles.map(f => 
            `${f.file.name}: ${f.errors.join(', ')}`
        );
        
        this.showError(`File validation failed:\n${errorMessages.join('\n')}`);
    }

    trackFileUpload(fileObj, status, error = null) {
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackEvent('file_upload', {
                category: fileObj.category,
                fileType: fileObj.file.type,
                fileSize: fileObj.file.size,
                status: status,
                error: error?.message
            });
        }
    }

    // Public API Methods
    static removeFile(zoneId, fileId) {
        if (window.PropertyHubFileUpload) {
            window.PropertyHubFileUpload.removeFile(zoneId, fileId);
        }
    }

    getUploadedFiles(zoneId) {
        const config = this.uploadZones.get(zoneId);
        return config ? config.files.filter(f => f.status === 'completed') : [];
    }

    getAllUploadedFiles() {
        const allFiles = [];
        for (const config of this.uploadZones.values()) {
            allFiles.push(...config.files.filter(f => f.status === 'completed'));
        }
        return allFiles;
    }

    clearZone(zoneId) {
        const config = this.uploadZones.get(zoneId);
        if (config) {
            // Cancel all uploads
            config.files.forEach(f => this.cancelUpload(f.id));
            
            // Clear files
            config.files = [];
            
            // Clear UI
            const filesContainer = document.getElementById(`files-${zoneId}`);
            if (filesContainer) {
                filesContainer.innerHTML = '';
            }
            
            // Show dropzone
            this.updateDropzoneVisibility(zoneId);
        }
    }

    // Batch Operations
    async uploadAll() {
        const allPendingFiles = [];
        
        for (const config of this.uploadZones.values()) {
            const pendingFiles = config.files.filter(f => f.status === 'pending' || f.status === 'processed');
            allPendingFiles.push(...pendingFiles.map(f => ({ zoneId: config.element.id, fileObj: f })));
        }

        for (const { zoneId, fileObj } of allPendingFiles) {
            this.addToUploadQueue(zoneId, fileObj);
        }

        this.processUploadQueue();
    }

    pauseAllUploads() {
        // Clear upload queue
        this.uploadQueue = [];
        
        // Cancel active uploads (would need xhr references in real implementation)
        this.activeUploads.clear();
    }

    retryFailedUploads() {
        for (const config of this.uploadZones.values()) {
            const failedFiles = config.files.filter(f => f.status === 'error');
            
            failedFiles.forEach(fileObj => {
                fileObj.status = 'pending';
                this.updateFileStatus(fileObj.id, 'Retrying...');
                this.addToUploadQueue(config.element.id, fileObj);
            });
        }

        this.processUploadQueue();
    }
}

// Initialize PropertyHub File Upload when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    if (document.querySelector('.file-upload-zone')) {
        window.PropertyHubFileUpload = new PropertyHubFileUpload();
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubFileUpload;
}
