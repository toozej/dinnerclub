/**
 * Dinnerclub Enhanced UI JavaScript
 * Provides responsive design, conditional form management, search interface, and smooth transitions
 */

// Global state management
window.DinnerClub = {
    isSearchActive: false,
    isMobile: window.innerWidth <= 768,
    tooltipTimeout: null,
    searchDebounceTimeout: null
};

/**
 * Conditional Form Field Management
 * Handles showing/hiding fields based on visited status in entry forms
 */
class ConditionalFormManager {
    constructor() {
        this.visitedCheckbox = document.getElementById('visited');
        this.visitedFields = document.getElementById('visitedFields');
        this.unvisitedFields = document.getElementById('unvisitedFields');
        this.lookingForwardTo = document.getElementById('lookingforwardto');

        if (this.visitedCheckbox) {
            this.init();
        }
    }

    init() {
        // Add change event listener with smooth transitions
        this.visitedCheckbox.addEventListener('change', () => {
            this.toggleFields();
        });

        // Initialize form state on page load
        this.toggleFields();

        // Add CSS classes for smooth transitions if not already present
        this.addTransitionStyles();
    }

    toggleFields() {
        const isVisited = this.visitedCheckbox.checked;

        if (isVisited) {
            this.showVisitedFields();
            this.hideUnvisitedFields();
            if (this.lookingForwardTo) {
                this.lookingForwardTo.removeAttribute('required');
            }
        } else {
            this.hideVisitedFields();
            this.showUnvisitedFields();
            if (this.lookingForwardTo) {
                this.lookingForwardTo.setAttribute('required', 'required');
            }
        }
    }

    showVisitedFields() {
        if (this.visitedFields) {
            this.visitedFields.classList.add('show');
            this.visitedFields.style.maxHeight = this.visitedFields.scrollHeight + 'px';
            this.visitedFields.style.opacity = '1';
        }
    }

    hideVisitedFields() {
        if (this.visitedFields) {
            this.visitedFields.classList.remove('show');
            this.visitedFields.style.maxHeight = '0';
            this.visitedFields.style.opacity = '0';
        }
    }

    showUnvisitedFields() {
        if (this.unvisitedFields) {
            this.unvisitedFields.classList.add('show');
            this.unvisitedFields.style.maxHeight = this.unvisitedFields.scrollHeight + 'px';
            this.unvisitedFields.style.opacity = '1';
        }
    }

    hideUnvisitedFields() {
        if (this.unvisitedFields) {
            this.unvisitedFields.classList.remove('show');
            this.unvisitedFields.style.maxHeight = '0';
            this.unvisitedFields.style.opacity = '0';
        }
    }

    addTransitionStyles() {
        const style = document.createElement('style');
        style.textContent = `
            .conditional-fields {
                overflow: hidden;
                transition: max-height 0.3s ease-in-out, opacity 0.3s ease-in-out;
                max-height: 0;
                opacity: 0;
            }
            
            .conditional-fields.show {
                max-height: none;
                opacity: 1;
            }
        `;
        document.head.appendChild(style);
    }
}

/**
 * Search Interface Manager
 * Handles search functionality with smooth transitions and state management
 */
class SearchManager {
    constructor() {
        this.searchInput = document.getElementById('searchInput');
        this.searchButton = document.getElementById('searchButton');
        this.clearButton = document.getElementById('clearSearch');
        this.defaultContent = document.getElementById('defaultContent');
        this.searchContent = document.getElementById('searchContent');

        if (this.searchInput && this.searchButton) {
            this.init();
        }
    }

    init() {
        // Event listeners
        this.searchButton.addEventListener('click', () => this.performSearch());
        if (this.clearButton) {
            this.clearButton.addEventListener('click', () => this.clearSearch());
        }

        // Enter key support with debouncing
        this.searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.performSearch();
            }
        });

        // Real-time search suggestions (debounced)
        this.searchInput.addEventListener('input', () => {
            clearTimeout(window.DinnerClub.searchDebounceTimeout);
            window.DinnerClub.searchDebounceTimeout = setTimeout(() => {
                this.showSearchSuggestions();
            }, 300);
        });

        // Escape key to clear search
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && window.DinnerClub.isSearchActive) {
                this.clearSearch();
            }
        });

        // Focus management
        this.searchInput.addEventListener('focus', () => {
            this.searchInput.parentElement.classList.add('focused');
        });

        this.searchInput.addEventListener('blur', () => {
            setTimeout(() => {
                this.searchInput.parentElement.classList.remove('focused');
            }, 200);
        });
    }

    async performSearch() {
        const query = this.searchInput.value.trim();
        if (!query) return;

        try {
            // Show loading state with smooth transition
            this.showLoadingState();
            this.showSearchResults();

            // Perform search request
            const response = await fetch(`/search?q=${encodeURIComponent(query)}`, {
                method: 'GET',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error('Search request failed');
            }

            const data = await response.json();

            // Display results with animation
            this.displaySearchResults(data, query);

        } catch (error) {
            console.error('Search error:', error);
            this.showErrorState('Search failed. Please try again.');
        }
    }

    showLoadingState() {
        if (this.searchContent) {
            this.searchContent.innerHTML = `
                <div class="loading">
                    <div class="loading-spinner"></div>
                    <span>Searching...</span>
                </div>
            `;
        }
    }

    showErrorState(message) {
        if (this.searchContent) {
            this.searchContent.innerHTML = `
                <div class="error">
                    <strong>Error:</strong> ${message}
                </div>
            `;
        }
    }

    displaySearchResults(data, query) {
        if (!this.searchContent) return;

        const totalResults = (data.entries?.length || 0) + (data.restaurants?.length || 0);

        let html = `
            <div class="search-results-container fade-in">
                <div class="search-info">
                    <h3>Search Results for: "${query}"</h3>
                    <p class="result-count">Found ${totalResults} total results</p>
                </div>
        `;

        if (totalResults > 0) {
            // Display entries
            if (data.entries?.length > 0) {
                html += this.renderEntryResults(data.entries);
            }

            // Display restaurants
            if (data.restaurants?.length > 0) {
                html += this.renderRestaurantResults(data.restaurants);
            }
        } else {
            html += this.renderNoResults();
        }

        html += '</div>';
        this.searchContent.innerHTML = html;

        // Trigger fade-in animation
        setTimeout(() => {
            const container = this.searchContent.querySelector('.search-results-container');
            if (container) {
                container.classList.add('visible');
            }
        }, 50);
    }

    renderEntryResults(entries) {
        let html = `
            <div class="search-results">
                <h3>Entry Results (${entries.length})</h3>
        `;

        entries.forEach((entry, index) => {
            html += `
                <section class="search-result-item" style="animation-delay: ${index * 0.1}s">
                    <h4><a href="/entries/${entry.ID}">${this.escapeHtml(entry.Name)}</a></h4>
                    <div class="result-details">
                        <span class="submitter">Submitter: <a href="/entries?author=${encodeURIComponent(entry.Submitter)}">${this.escapeHtml(entry.Submitter)}</a></span><br>
                        <span class="date">Date: ${new Date(entry.CreatedAt).toLocaleDateString()}</span><br>
                        <span class="visited">Visited: ${entry.Visited ? 'Yes' : 'No'}</span><br>
                        <span class="location">Location: ${this.escapeHtml(entry.Location)}</span><br>
                        <span class="cuisine">Cuisine: ${this.escapeHtml(entry.Cuisine)}</span><br>
                        ${entry.LookingForwardTo ? `<span class="looking-forward">Looking forward to: ${this.escapeHtml(entry.LookingForwardTo)}</span><br>` : ''}
                        ${entry.Comments ? `<span class="comments">Comments: ${this.escapeHtml(entry.Comments)}</span><br>` : ''}
                    </div>
                </section>
            `;
        });

        html += '</div>';
        return html;
    }

    renderRestaurantResults(restaurants) {
        let html = `
            <div class="search-results">
                <h3>Restaurant Results (${restaurants.length})</h3>
        `;

        restaurants.forEach((restaurant, index) => {
            html += `
                <section class="search-result-item" style="animation-delay: ${(index + 10) * 0.1}s">
                    <h4><a href="/restaurants/name/${encodeURIComponent(restaurant.Name)}">${this.escapeHtml(restaurant.Name)}</a></h4>
                    <div class="result-details">
                        <span class="location">Location: ${this.escapeHtml(restaurant.Location)}</span><br>
                        <span class="cuisine">Cuisine: ${this.escapeHtml(restaurant.Cuisine)}</span><br>
                        ${restaurant.EntryCount ? `<span class="entry-count">Entries: ${restaurant.EntryCount}</span><br>` : ''}
                    </div>
                </section>
            `;
        });

        html += '</div>';
        return html;
    }

    renderNoResults() {
        return `
            <div class="no-results">
                <h3>No results found</h3>
                <p>Try adjusting your search terms or check for typos.</p>
                <p><strong>Search tips:</strong></p>
                <ul>
                    <li>Use "wishlist" to find places you want to visit</li>
                    <li>Use "author:username" to find entries by a specific user</li>
                    <li>Search by cuisine type or location</li>
                </ul>
            </div>
        `;
    }

    showSearchResults() {
        if (this.defaultContent && this.searchContent) {
            this.defaultContent.style.display = 'none';
            this.searchContent.style.display = 'block';
            if (this.clearButton) {
                this.clearButton.style.display = 'inline-block';
            }
            window.DinnerClub.isSearchActive = true;
        }
    }

    clearSearch() {
        if (this.defaultContent && this.searchContent) {
            this.searchContent.style.display = 'none';
            this.defaultContent.style.display = 'block';
            if (this.clearButton) {
                this.clearButton.style.display = 'none';
            }
            this.searchInput.value = '';
            window.DinnerClub.isSearchActive = false;
        }
    }

    showSearchSuggestions() {
        // Placeholder for future search suggestions functionality
        // Could show dropdown with recent searches or popular terms
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

/**
 * Tooltip System
 * Provides contextual help for form fields and UI elements
 */
class TooltipManager {
    constructor() {
        this.tooltips = new Map();
        this.activeTooltip = null;
        this.init();
    }

    init() {
        // Initialize tooltips for form fields
        this.initFormTooltips();

        // Add global event listeners
        document.addEventListener('mouseover', (e) => this.handleMouseOver(e));
        document.addEventListener('mouseout', (e) => this.handleMouseOut(e));
        document.addEventListener('focus', (e) => this.handleFocus(e), true);
        document.addEventListener('blur', (e) => this.handleBlur(e), true);
    }

    initFormTooltips() {
        const tooltipData = {
            'name': 'Enter the restaurant name',
            'location': 'Select the general area/neighborhood',
            'cuisine': 'Type of cuisine (e.g., Italian, Thai, American)',
            'visited': 'Check if you have already visited this restaurant',
            'lookingforwardto': 'What dish or experience are you most excited to try?',
            'ordered': 'List items you ordered, separated by commas',
            'ordernext': 'Items you want to try next time, separated by commas',
            'donotorder': 'Items to avoid, separated by commas',
            'mealservice': 'Type of meal (breakfast, lunch, dinner, etc.)',
            'foodrating': 'Rate the food quality (1-5 stars)',
            'ambiencerating': 'Rate the atmosphere and ambience (1-5 stars)',
            'valuerating': 'Rate the value for money (1-5 stars)',
            'comments': 'Additional thoughts or details about your experience',
            'closed': 'Check if the restaurant is permanently closed'
        };

        Object.entries(tooltipData).forEach(([fieldId, text]) => {
            this.tooltips.set(fieldId, text);
        });
    }

    handleMouseOver(e) {
        const element = e.target;
        if (this.shouldShowTooltip(element)) {
            clearTimeout(window.DinnerClub.tooltipTimeout);
            window.DinnerClub.tooltipTimeout = setTimeout(() => {
                this.showTooltip(element);
            }, 500);
        }
    }

    handleMouseOut(e) {
        clearTimeout(window.DinnerClub.tooltipTimeout);
        this.hideTooltip();
    }

    handleFocus(e) {
        const element = e.target;
        if (this.shouldShowTooltip(element)) {
            this.showTooltip(element);
        }
    }

    handleBlur(e) {
        this.hideTooltip();
    }

    shouldShowTooltip(element) {
        // Check if element exists and is a valid DOM element
        if (!element || !element.tagName || typeof element.hasAttribute !== 'function') {
            return false;
        }

        return element.tagName === 'INPUT' ||
            element.tagName === 'SELECT' ||
            element.tagName === 'TEXTAREA' ||
            element.hasAttribute('data-tooltip');
    }

    showTooltip(element) {
        const tooltipText = this.getTooltipText(element);
        if (!tooltipText) return;

        this.hideTooltip(); // Hide any existing tooltip

        const tooltip = document.createElement('div');
        tooltip.className = 'tooltip';
        tooltip.textContent = tooltipText;

        document.body.appendChild(tooltip);
        this.activeTooltip = tooltip;

        // Position tooltip
        this.positionTooltip(tooltip, element);

        // Show with animation
        setTimeout(() => {
            tooltip.classList.add('visible');
        }, 10);
    }

    hideTooltip() {
        if (this.activeTooltip) {
            this.activeTooltip.classList.remove('visible');
            setTimeout(() => {
                if (this.activeTooltip && this.activeTooltip.parentNode) {
                    this.activeTooltip.parentNode.removeChild(this.activeTooltip);
                }
                this.activeTooltip = null;
            }, 200);
        }
    }

    getTooltipText(element) {
        // Check if element exists and is a valid DOM element
        if (!element || !element.tagName || typeof element.hasAttribute !== 'function') {
            return null;
        }

        // Check for custom tooltip attribute first
        if (element.hasAttribute('data-tooltip')) {
            return element.getAttribute('data-tooltip');
        }

        // Check for predefined tooltips
        if (element.id && this.tooltips.has(element.id)) {
            return this.tooltips.get(element.id);
        }

        // Check for label-based tooltips
        if (element.id) {
            const label = document.querySelector(`label[for="${element.id}"]`);
            if (label && typeof label.hasAttribute === 'function' && label.hasAttribute('data-tooltip')) {
                return label.getAttribute('data-tooltip');
            }
        }

        return null;
    }

    positionTooltip(tooltip, element) {
        const rect = element.getBoundingClientRect();
        const tooltipRect = tooltip.getBoundingClientRect();

        let top = rect.bottom + window.scrollY + 5;
        let left = rect.left + window.scrollX;

        // Adjust if tooltip would go off screen
        if (left + tooltipRect.width > window.innerWidth) {
            left = window.innerWidth - tooltipRect.width - 10;
        }

        if (top + tooltipRect.height > window.innerHeight + window.scrollY) {
            top = rect.top + window.scrollY - tooltipRect.height - 5;
        }

        tooltip.style.top = `${top}px`;
        tooltip.style.left = `${left}px`;
    }
}

/**
 * Responsive Design Manager
 * Handles mobile-specific behaviors and responsive adjustments
 */
class ResponsiveManager {
    constructor() {
        this.breakpoints = {
            mobile: 768,
            tablet: 1024
        };

        this.init();
    }

    init() {
        // Initial setup
        this.updateViewportInfo();

        // Listen for resize events
        window.addEventListener('resize', () => {
            clearTimeout(this.resizeTimeout);
            this.resizeTimeout = setTimeout(() => {
                this.updateViewportInfo();
                this.handleResize();
            }, 250);
        });

        // Handle orientation changes on mobile
        window.addEventListener('orientationchange', () => {
            setTimeout(() => {
                this.updateViewportInfo();
                this.handleOrientationChange();
            }, 500);
        });
    }

    updateViewportInfo() {
        const oldIsMobile = window.DinnerClub.isMobile;
        window.DinnerClub.isMobile = window.innerWidth <= this.breakpoints.mobile;

        // Update body class for CSS targeting
        document.body.classList.toggle('mobile', window.DinnerClub.isMobile);
        document.body.classList.toggle('desktop', !window.DinnerClub.isMobile);

        // Trigger mobile state change event if needed
        if (oldIsMobile !== window.DinnerClub.isMobile) {
            this.handleMobileStateChange();
        }
    }

    handleResize() {
        // Adjust search header height on mobile
        if (window.DinnerClub.isMobile) {
            this.adjustMobileLayout();
        }

        // Reposition any active tooltips
        if (window.DinnerClub.tooltipManager && window.DinnerClub.tooltipManager.activeTooltip) {
            window.DinnerClub.tooltipManager.hideTooltip();
        }
    }

    handleOrientationChange() {
        // Force layout recalculation after orientation change
        this.adjustMobileLayout();
    }

    handleMobileStateChange() {
        if (window.DinnerClub.isMobile) {
            this.enableMobileOptimizations();
        } else {
            this.disableMobileOptimizations();
        }
    }

    adjustMobileLayout() {
        const searchHeader = document.querySelector('.search-header');
        const contentArea = document.querySelector('.content-area');

        if (searchHeader && contentArea) {
            const headerHeight = searchHeader.offsetHeight;
            contentArea.style.top = `${headerHeight}px`;
        }
    }

    enableMobileOptimizations() {
        // Add mobile-specific behaviors
        this.addTouchOptimizations();
        this.adjustFormLayout();
    }

    disableMobileOptimizations() {
        // Remove mobile-specific behaviors
        this.removeTouchOptimizations();
    }

    addTouchOptimizations() {
        // Increase touch targets
        document.body.classList.add('touch-optimized');

        // Add touch-friendly button spacing
        const buttons = document.querySelectorAll('.btn, .nav-button, .add-button');
        buttons.forEach(button => {
            button.style.minHeight = '44px';
            button.style.minWidth = '44px';
        });
    }

    removeTouchOptimizations() {
        document.body.classList.remove('touch-optimized');
    }

    adjustFormLayout() {
        // Stack form elements vertically on mobile
        const forms = document.querySelectorAll('form');
        forms.forEach(form => {
            form.classList.add('mobile-form');
        });
    }
}

/**
 * Animation and Transition Manager
 * Provides smooth animations and transitions throughout the UI
 */
class AnimationManager {
    constructor() {
        this.init();
    }

    init() {
        // Add CSS for animations
        this.addAnimationStyles();

        // Initialize page load animations
        this.initPageAnimations();
    }

    addAnimationStyles() {
        const style = document.createElement('style');
        style.textContent = `
            /* Fade in animation */
            .fade-in {
                opacity: 0;
                transform: translateY(20px);
                transition: opacity 0.3s ease-out, transform 0.3s ease-out;
            }
            
            .fade-in.visible {
                opacity: 1;
                transform: translateY(0);
            }
            
            /* Search result animations */
            .search-result-item {
                opacity: 0;
                transform: translateX(-20px);
                animation: slideInLeft 0.3s ease-out forwards;
            }
            
            @keyframes slideInLeft {
                to {
                    opacity: 1;
                    transform: translateX(0);
                }
            }
            
            /* Loading spinner */
            .loading-spinner {
                display: inline-block;
                width: 20px;
                height: 20px;
                border: 3px solid #f3f3f3;
                border-top: 3px solid #007bff;
                border-radius: 50%;
                animation: spin 1s linear infinite;
                margin-right: 10px;
            }
            
            @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }
            
            /* Button hover animations */
            .btn, .nav-button, .add-button {
                transition: all 0.15s ease-in-out;
            }
            
            /* Form field focus animations */
            input:focus, select:focus, textarea:focus {
                transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
            }
            
            /* Tooltip animations */
            .tooltip {
                position: absolute;
                background-color: #333;
                color: white;
                padding: 8px 12px;
                border-radius: 4px;
                font-size: 0.85rem;
                z-index: 10000;
                opacity: 0;
                transform: translateY(-5px);
                transition: opacity 0.2s ease-out, transform 0.2s ease-out;
                pointer-events: none;
                max-width: 250px;
                word-wrap: break-word;
            }
            
            .tooltip.visible {
                opacity: 1;
                transform: translateY(0);
            }
            
            .tooltip::after {
                content: '';
                position: absolute;
                top: -5px;
                left: 50%;
                transform: translateX(-50%);
                border-left: 5px solid transparent;
                border-right: 5px solid transparent;
                border-bottom: 5px solid #333;
            }
            
            /* Search box focus animation */
            .search-box.focused {
                transform: scale(1.02);
                transition: transform 0.2s ease-out;
            }
            
            /* Mobile-specific animations */
            @media (max-width: 768px) {
                .fade-in {
                    transform: translateY(10px);
                }
                
                .search-result-item {
                    transform: translateY(10px);
                    animation: slideInUp 0.3s ease-out forwards;
                }
                
                @keyframes slideInUp {
                    to {
                        opacity: 1;
                        transform: translateY(0);
                    }
                }
            }
        `;
        document.head.appendChild(style);
    }

    initPageAnimations() {
        // Animate elements on page load
        document.addEventListener('DOMContentLoaded', () => {
            const animatedElements = document.querySelectorAll('.search-result-item, .entry-item, .restaurant-item');
            animatedElements.forEach((element, index) => {
                element.style.animationDelay = `${index * 0.1}s`;
            });
        });
    }
}

/**
 * Auto-Complete Manager
 * Provides auto-complete functionality for form fields based on existing data
 */
class AutoCompleteManager {
    constructor() {
        this.cache = new Map();
        this.prePopulatedData = new Map();
        this.activeDropdown = null;
        this.debounceTimeout = null;
        this.init();
    }

    init() {
        // Pre-populate data for all fields
        this.prePopulateAllFields();

        // Initialize auto-complete for specific fields
        this.initAutoCompleteFields();

        // Add global click listener to close dropdowns
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.autocomplete-container')) {
                this.hideDropdown();
            }
        });

        // Add escape key listener
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.activeDropdown) {
                this.hideDropdown();
            }
        });
    }

    async prePopulateAllFields() {
        const autoCompleteFields = [
            { id: 'cuisine', endpoint: '/api/suggestions/cuisine' },
            { id: 'name', endpoint: '/api/suggestions/restaurant-names' },
            { id: 'mealservice', endpoint: '/api/suggestions/meal-service' },
            { id: 'location', endpoint: '/api/suggestions/locations' },
            { id: 'ordered', endpoint: '/api/suggestions/dishes', params: { type: 'ordered' } },
            { id: 'ordernext', endpoint: '/api/suggestions/dishes', params: { type: 'ordernext' } },
            { id: 'donotorder', endpoint: '/api/suggestions/dishes', params: { type: 'donotorder' } }
        ];

        // Pre-populate data for all fields
        for (const field of autoCompleteFields) {
            try {
                const suggestions = await this.fetchSuggestions(field.endpoint, '', field.params || {});
                this.prePopulatedData.set(field.id, suggestions);
            } catch (error) {
                console.warn(`Failed to pre-populate ${field.id}:`, error);
            }
        }
    }

    initAutoCompleteFields() {
        const autoCompleteFields = [
            { id: 'cuisine', endpoint: '/api/suggestions/cuisine' },
            { id: 'name', endpoint: '/api/suggestions/restaurant-names' },
            { id: 'mealservice', endpoint: '/api/suggestions/meal-service' },
            { id: 'location', endpoint: '/api/suggestions/locations' },
            { id: 'ordered', endpoint: '/api/suggestions/dishes', params: { type: 'ordered' } },
            { id: 'ordernext', endpoint: '/api/suggestions/dishes', params: { type: 'ordernext' } },
            { id: 'donotorder', endpoint: '/api/suggestions/dishes', params: { type: 'donotorder' } }
        ];

        autoCompleteFields.forEach(field => {
            const element = document.getElementById(field.id);
            if (element) {
                this.setupAutoComplete(element, field.endpoint, field.params);
            }
        });
    }

    setupAutoComplete(element, endpoint, params = {}) {
        // Wrap the input in a container for positioning
        const container = this.createAutoCompleteContainer(element);

        // Add CSS class to indicate auto-complete is enabled
        element.classList.add('autocomplete-enabled');

        // Add event listeners
        element.addEventListener('input', (e) => {
            clearTimeout(this.debounceTimeout);
            this.debounceTimeout = setTimeout(() => {
                this.handleInput(e.target, endpoint, params);
            }, 300);
        });

        element.addEventListener('focus', (e) => {
            // Mark container as active
            container.classList.add('active');
            // Always show suggestions on focus - either filtered by current text or all pre-populated
            this.handleInput(e.target, endpoint, params);
        });

        element.addEventListener('blur', (e) => {
            // Remove active state after a short delay to allow for clicks
            setTimeout(() => {
                container.classList.remove('active');
            }, 150);
        });

        element.addEventListener('keydown', (e) => {
            this.handleKeyDown(e);
        });
    }

    createAutoCompleteContainer(element) {
        // Check if already wrapped
        if (element.parentElement.classList.contains('autocomplete-container')) {
            return element.parentElement;
        }

        const container = document.createElement('div');
        container.className = 'autocomplete-container';
        element.parentNode.insertBefore(container, element);
        container.appendChild(element);

        return container;
    }

    async handleInput(element, endpoint, params = {}) {
        const query = element.value.trim();
        const fieldId = element.id;

        // If no query, show pre-populated suggestions
        if (query.length === 0) {
            const prePopulated = this.prePopulatedData.get(fieldId);
            if (prePopulated && prePopulated.length > 0) {
                this.showSuggestions(element, prePopulated.slice(0, 10)); // Show top 10
            } else {
                this.hideDropdown();
            }
            return;
        }

        try {
            let suggestions;

            // For short queries, try to filter pre-populated data first
            if (query.length <= 2) {
                const prePopulated = this.prePopulatedData.get(fieldId);
                if (prePopulated) {
                    suggestions = this.filterSuggestions(prePopulated, query);

                    // If we have enough results from pre-populated data, use them
                    if (suggestions.length >= 3) {
                        this.showSuggestions(element, suggestions.slice(0, 10));
                        return;
                    }
                }
            }

            // Fetch fresh suggestions from API
            suggestions = await this.fetchSuggestions(endpoint, query, params);
            this.showSuggestions(element, suggestions);
        } catch (error) {
            console.error('Auto-complete error:', error);

            // Fallback to pre-populated data if API fails
            const prePopulated = this.prePopulatedData.get(fieldId);
            if (prePopulated) {
                const filtered = this.filterSuggestions(prePopulated, query);
                this.showSuggestions(element, filtered.slice(0, 10));
            } else {
                this.hideDropdown();
            }
        }
    }

    async fetchSuggestions(endpoint, query, params = {}) {
        // Create cache key
        const cacheKey = `${endpoint}:${query}:${JSON.stringify(params)}`;

        // Check cache first
        if (this.cache.has(cacheKey)) {
            return this.cache.get(cacheKey);
        }

        // Build URL with parameters
        const url = new URL(endpoint, window.location.origin);
        if (query) {
            url.searchParams.set('q', query);
        }
        Object.entries(params).forEach(([key, value]) => {
            url.searchParams.set(key, value);
        });

        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            // Add timeout to prevent hanging requests
            signal: AbortSignal.timeout(10000)
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();
        const suggestions = data.data?.suggestions || [];

        // Cache the results for 5 minutes
        this.cache.set(cacheKey, suggestions);
        setTimeout(() => {
            this.cache.delete(cacheKey);
        }, 5 * 60 * 1000);

        return suggestions;
    }

    filterSuggestions(suggestions, query) {
        if (!query || query.length === 0) {
            return suggestions;
        }

        const lowerQuery = query.toLowerCase();

        return suggestions.filter(suggestion => {
            const value = suggestion.value.toLowerCase();
            return value.includes(lowerQuery);
        }).sort((a, b) => {
            const aValue = a.value.toLowerCase();
            const bValue = b.value.toLowerCase();

            // Prioritize exact matches at the beginning
            const aStartsWith = aValue.startsWith(lowerQuery);
            const bStartsWith = bValue.startsWith(lowerQuery);

            if (aStartsWith && !bStartsWith) return -1;
            if (!aStartsWith && bStartsWith) return 1;

            // Then sort by count (popularity)
            if (a.count !== b.count) {
                return (b.count || 0) - (a.count || 0);
            }

            // Finally sort alphabetically
            return aValue.localeCompare(bValue);
        });
    }

    showSuggestions(element, suggestions) {
        this.hideDropdown();

        const container = element.closest('.autocomplete-container');
        const dropdown = document.createElement('div');
        dropdown.className = 'autocomplete-dropdown';

        if (!suggestions || suggestions.length === 0) {
            // Show "no results" message if user has typed something
            if (element.value.trim().length > 0) {
                const noResults = document.createElement('div');
                noResults.className = 'autocomplete-no-results';
                noResults.textContent = 'No suggestions found';
                dropdown.appendChild(noResults);
            } else {
                return; // Don't show dropdown if no query and no suggestions
            }
        } else {
            // Limit suggestions to prevent overwhelming the user
            const limitedSuggestions = suggestions.slice(0, 15);

            limitedSuggestions.forEach((suggestion, index) => {
                const item = document.createElement('div');
                item.className = 'autocomplete-item';
                item.setAttribute('data-index', index);

                // Create suggestion content with count if available
                const content = document.createElement('span');
                content.className = 'suggestion-text';
                content.textContent = suggestion.value;
                item.appendChild(content);

                if (suggestion.count && suggestion.count > 1) {
                    const count = document.createElement('span');
                    count.className = 'suggestion-count';
                    count.textContent = suggestion.count;
                    item.appendChild(count);
                }

                // Add click handler
                item.addEventListener('click', (e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    this.selectSuggestion(element, suggestion.value);
                });

                dropdown.appendChild(item);
            });
        }

        container.appendChild(dropdown);
        this.activeDropdown = dropdown;

        // Position dropdown
        this.positionDropdown(dropdown, element);

        // Show with animation
        setTimeout(() => {
            dropdown.classList.add('visible');
        }, 10);
    }

    hideDropdown() {
        if (this.activeDropdown) {
            this.activeDropdown.classList.remove('visible');
            setTimeout(() => {
                if (this.activeDropdown && this.activeDropdown.parentNode) {
                    this.activeDropdown.parentNode.removeChild(this.activeDropdown);
                }
                this.activeDropdown = null;
            }, 150);
        }
    }

    selectSuggestion(element, value) {
        element.value = value;
        this.hideDropdown();

        // Trigger change event
        const event = new Event('change', { bubbles: true });
        element.dispatchEvent(event);

        // Focus back to the input
        element.focus();
    }

    handleKeyDown(e) {
        if (!this.activeDropdown) return;

        const items = this.activeDropdown.querySelectorAll('.autocomplete-item');
        const currentSelected = this.activeDropdown.querySelector('.autocomplete-item.selected');
        let selectedIndex = currentSelected ? parseInt(currentSelected.getAttribute('data-index')) : -1;

        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault();
                selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
                this.updateSelection(items, selectedIndex);
                break;

            case 'ArrowUp':
                e.preventDefault();
                selectedIndex = Math.max(selectedIndex - 1, 0);
                this.updateSelection(items, selectedIndex);
                break;

            case 'Enter':
                e.preventDefault();
                if (currentSelected) {
                    const text = currentSelected.querySelector('.suggestion-text').textContent;
                    this.selectSuggestion(e.target, text);
                }
                break;

            case 'Tab':
                if (currentSelected) {
                    e.preventDefault();
                    const text = currentSelected.querySelector('.suggestion-text').textContent;
                    this.selectSuggestion(e.target, text);
                }
                break;
        }
    }

    updateSelection(items, selectedIndex) {
        items.forEach((item, index) => {
            item.classList.toggle('selected', index === selectedIndex);
        });
    }

    positionDropdown(dropdown, element) {
        const rect = element.getBoundingClientRect();
        const dropdownRect = dropdown.getBoundingClientRect();

        // Position below the input
        dropdown.style.top = '100%';
        dropdown.style.left = '0';
        dropdown.style.width = '100%';
        dropdown.style.minWidth = `${rect.width}px`;

        // Adjust if dropdown would go off screen
        const spaceBelow = window.innerHeight - rect.bottom;
        const spaceAbove = rect.top;

        if (dropdownRect.height > spaceBelow && spaceAbove > spaceBelow) {
            dropdown.style.top = 'auto';
            dropdown.style.bottom = '100%';
        }
    }
}

/**
 * Initialize all managers when DOM is ready
 */
document.addEventListener('DOMContentLoaded', function () {
    // Initialize all UI managers
    window.DinnerClub.conditionalFormManager = new ConditionalFormManager();
    window.DinnerClub.searchManager = new SearchManager();
    window.DinnerClub.tooltipManager = new TooltipManager();
    window.DinnerClub.responsiveManager = new ResponsiveManager();
    window.DinnerClub.animationManager = new AnimationManager();
    window.DinnerClub.autoCompleteManager = new AutoCompleteManager();

    console.log('DinnerClub Enhanced UI initialized');
});

// Legacy support for existing inline scripts
function toggleConditionalFields() {
    if (window.DinnerClub.conditionalFormManager) {
        window.DinnerClub.conditionalFormManager.toggleFields();
    }
}