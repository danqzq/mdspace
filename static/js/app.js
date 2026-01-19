// mdspace - Main Application JavaScript

// Configure marked for GFM support
marked.setOptions({
    gfm: true,
    breaks: true,
    highlight: function (code, lang) {
        if (lang && hljs.getLanguage(lang)) {
            return hljs.highlight(code, { language: lang }).value;
        }
        return hljs.highlightAuto(code).value;
    }
});

// DOM Elements
const dropZone = document.getElementById('drop-zone');
const fileInput = document.getElementById('file-input');
const markdownInput = document.getElementById('markdown-input');
const previewContent = document.getElementById('preview-content');
const clearBtn = document.getElementById('clear-btn');
const shareBtn = document.getElementById('share-btn');
const shareModal = document.getElementById('share-modal');
const shareLink = document.getElementById('share-link');
const copyBtn = document.getElementById('copy-btn');
const viewBtn = document.getElementById('view-btn');
const closeModalBtn = document.getElementById('close-modal-btn');
const expiresInfo = document.getElementById('expires-info');
const filesCount = document.getElementById('files-count');
const filesLimit = document.getElementById('files-limit');

let currentShareUrl = '';

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadUserStats();
    setupEventListeners();
});

async function loadUserStats() {
    try {
        const response = await fetch('/api/user/stats');
        if (response.ok) {
            const data = await response.json();
            filesCount.textContent = data.files_count;
            filesLimit.textContent = data.files_limit;
        }
    } catch (error) {
        console.error('Failed to load user stats:', error);
    }
}

// Set up event listeners
function setupEventListeners() {
    dropZone.addEventListener('click', () => fileInput.click());

    dropZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropZone.classList.add('dragover');
    });

    dropZone.addEventListener('dragleave', () => {
        dropZone.classList.remove('dragover');
    });

    dropZone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropZone.classList.remove('dragover');

        const files = e.dataTransfer.files;
        if (files.length > 0) {
            handleFile(files[0]);
        }
    });

    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            handleFile(e.target.files[0]);
        }
    });

    let debounceTimer;
    markdownInput.addEventListener('input', () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            updatePreview();
        }, 150);
    });

    clearBtn.addEventListener('click', () => {
        markdownInput.value = '';
        updatePreview();
    });

    shareBtn.addEventListener('click', createShareLink);

    copyBtn.addEventListener('click', copyShareLink);
    viewBtn.addEventListener('click', () => {
        window.open(currentShareUrl, '_blank');
    });
    closeModalBtn.addEventListener('click', closeModal);

    shareModal.addEventListener('click', (e) => {
        if (e.target === shareModal) {
            closeModal();
        }
    });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && shareModal.classList.contains('active')) {
            closeModal();
        }
    });
}

function handleFile(file) {
    if (!file.name.match(/\.(md|markdown)$/i) && file.type !== 'text/markdown') {
        showToast('Please upload a markdown file (.md)', 'error');
        return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
        markdownInput.value = e.target.result;
        updatePreview();
        showToast('File loaded successfully', 'success');
    };
    reader.onerror = () => {
        showToast('Failed to read file', 'error');
    };
    reader.readAsText(file);
}

function updatePreview() {
    const content = markdownInput.value.trim();

    if (!content) {
        previewContent.innerHTML = '<div class="preview-placeholder">Your markdown preview will appear here...</div>';
        shareBtn.disabled = true;
        return;
    }

    try {
        const html = marked.parse(content);
        previewContent.innerHTML = `<div class="markdown-body">${html}</div>`;

        previewContent.querySelectorAll('pre code').forEach((block) => {
            hljs.highlightElement(block);
        });

        shareBtn.disabled = false;
    } catch (error) {
        previewContent.innerHTML = '<div class="preview-placeholder">Error rendering markdown</div>';
        shareBtn.disabled = true;
    }
}

async function createShareLink() {
    const content = markdownInput.value.trim();

    if (!content) {
        showToast('Please add some markdown content', 'error');
        return;
    }

    shareBtn.disabled = true;
    shareBtn.innerHTML = '<div class="spinner"></div> Creating...';

    try {
        const response = await fetch('/api/markdown', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ content })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to create share link');
        }

        loadUserStats();

        currentShareUrl = data.share_url;
        shareLink.value = data.share_url;

        const expiresAt = new Date(data.expires_at);
        expiresInfo.textContent = `This link will expire on ${expiresAt.toLocaleString()}`;

        shareModal.classList.add('active');
    } catch (error) {
        showToast(error.message, 'error');
    } finally {
        shareBtn.disabled = false;
        shareBtn.innerHTML = '<span>Create Share Link</span>';
    }
}

async function copyShareLink() {
    try {
        await navigator.clipboard.writeText(shareLink.value);
        copyBtn.textContent = 'Copied!';
        setTimeout(() => {
            copyBtn.textContent = 'Copy';
        }, 2000);
    } catch (error) {
        showToast('Failed to copy link', 'error');
    }
}

function closeModal() {
    shareModal.classList.remove('active');
    markdownInput.value = '';
    updatePreview();
}

function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;

    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}
