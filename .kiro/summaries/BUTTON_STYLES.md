# Button Style Guide

This document outlines the consistent button styling system implemented for the Dinnerclub application.

## Base Button Classes

### `.btn`
The base button class that provides core styling for all buttons.

### Button Variants

- `.btn-primary` - Primary action buttons (blue)
- `.btn-success` - Success/positive action buttons (green) 
- `.btn-secondary` - Secondary action buttons (gray)
- `.btn-danger` - Destructive action buttons (red)
- `.btn-outline-primary` - Outlined primary buttons
- `.btn-outline-secondary` - Outlined secondary buttons

### Button Sizes

- `.btn-sm` - Small buttons
- `.btn-lg` - Large buttons  
- `.btn-xl` - Extra large buttons
- `.btn-block` - Full width buttons

### Special Purpose Classes

- `.btn-nav` - Navigation menu buttons
- `.btn-add` - Add/create action buttons (green with + symbol)
- `.btn-form` - Form submission buttons
- `.btn-search` - Search-related buttons

## Usage Examples

### Navigation Buttons
```html
<a href="/" class="btn btn-nav">Home</a>
<button class="btn btn-nav" type="submit">Logout</button>
```

### Form Buttons
```html
<button type="submit" class="btn btn-primary btn-form">Submit</button>
<button type="submit" class="btn btn-success btn-form">Update</button>
<button type="submit" class="btn btn-danger btn-form">Delete</button>
```

### Action Buttons
```html
<a href="/entries/new" class="btn btn-add">+</a>
<button class="btn btn-primary">Edit</button>
<button class="btn btn-danger">Delete</button>
```

### Search Buttons
```html
<button class="btn btn-primary btn-search">Search</button>
<button class="btn btn-secondary btn-search">Clear</button>
```

## Legacy Support

The following legacy classes are maintained for backward compatibility:
- `.nav-button` - Maps to `.btn .btn-nav`
- `.add-button` - Maps to `.btn .btn-add`  
- `.submit-btn` - Maps to `.btn .btn-success .btn-form`

## Features

- Consistent hover and active states
- Focus indicators for accessibility
- Responsive design support
- Smooth transitions
- Disabled state styling
- Button groups support