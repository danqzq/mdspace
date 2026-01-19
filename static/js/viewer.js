// mdspace - Viewer Page JavaScript

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

// State
let markdownId = '';
let isOwner = false;
let selectedLine = null;
let comments = [];
let rawContent = '';

// DOM Elements
const loadingState = document.getElementById('loading-state');
const notFoundState = document.getElementById('not-found-state');
const contentState = document.getElementById('content-state');
const markdownContent = document.getElementById('markdown-content');
const viewCount = document.getElementById('view-count');
const expiresAt = document.getElementById('expires-at');
const deleteBtn = document.getElementById('delete-btn');
const copyLinkBtn = document.getElementById('copy-link-btn');
const downloadBtn = document.getElementById('download-btn');
const commentsList = document.getElementById('comments-list');
const commentCount = document.getElementById('comment-count');
const commentForm = document.getElementById('comment-form');
const commentLineNum = document.getElementById('comment-line-num');
const commentAuthor = document.getElementById('comment-author');
const commentText = document.getElementById('comment-text');
const submitCommentBtn = document.getElementById('submit-comment-btn');
const cancelCommentBtn = document.getElementById('cancel-comment-btn');

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    // Extract markdown ID from URL
    const pathParts = window.location.pathname.split('/');
    markdownId = pathParts[pathParts.length - 1];

    if (!markdownId) {
        showNotFound();
        return;
    }

    loadMarkdown();
    setupEventListeners();
});

// Set up event listeners
function setupEventListeners() {
    copyLinkBtn.addEventListener('click', copyCurrentLink);
    downloadBtn.addEventListener('click', downloadMarkdown);
    deleteBtn.addEventListener('click', deleteMarkdown);
    submitCommentBtn.addEventListener('click', submitComment);
    cancelCommentBtn.addEventListener('click', cancelComment);
}

async function loadMarkdown() {
    try {
        const response = await fetch(`/api/markdown/${markdownId}`);

        if (!response.ok) {
            if (response.status === 404) {
                showNotFound();
                return;
            }
            throw new Error('Failed to load markdown');
        }

        const data = await response.json();

        viewCount.textContent = data.views;

        const expires = new Date(data.expires_at);
        expiresAt.textContent = formatRelativeTime(expires);

        const firstLine = data.content.split('\n')[0].replace(/^#*\s*/, '');
        document.title = `${firstLine.substring(0, 50)} - mdspace`;

        isOwner = data.is_owner;
        if (isOwner) {
            deleteBtn.classList.remove('hidden');
        }

        // Store raw content for download
        rawContent = data.content;

        renderMarkdownWithLines(data.content);

        await loadComments();

        loadingState.classList.add('hidden');
        contentState.classList.remove('hidden');

    } catch (error) {
        console.error('Error loading markdown:', error);
        showToast('Failed to load markdown', 'error');
        showNotFound();
    }
}

function renderMarkdownWithLines(content) {
    const lines = content.split('\n');

    const lineHTML = lines.map((line, index) => {
        const lineNum = index + 1;
        const escapedLine = escapeHtml(line);
        return `
            <div class="code-line" data-line="${lineNum}">
                <span class="line-number" data-line="${lineNum}">${lineNum}</span>
                <span class="line-content">${escapedLine || ' '}</span>
            </div>
        `;
    }).join('');

    const renderedMarkdown = marked.parse(content);

    markdownContent.innerHTML = `
        <div class="view-tabs">
            <button class="view-tab active" data-view="rendered">Rendered</button>
            <button class="view-tab" data-view="source">Source (for comments)</button>
        </div>
        <div class="view-panel active" id="rendered-view">
            <div class="markdown-body">${renderedMarkdown}</div>
        </div>
        <div class="view-panel" id="source-view">
            <div class="line-numbered-content">${lineHTML}</div>
        </div>
    `;

    const style = document.createElement('style');
    style.textContent = `
        .view-tabs {
            display: flex;
            gap: 4px;
            margin-bottom: 16px;
            border-bottom: 1px solid var(--color-border-default);
            padding-bottom: 8px;
        }
        .view-tab {
            padding: 8px 16px;
            background: transparent;
            border: none;
            color: var(--color-text-secondary);
            cursor: pointer;
            border-radius: 6px 6px 0 0;
            font-size: 0.875rem;
            transition: all 0.15s;
        }
        .view-tab:hover {
            color: var(--color-text-primary);
            background: var(--color-bg-tertiary);
        }
        .view-tab.active {
            color: var(--color-text-primary);
            background: var(--color-bg-tertiary);
            border-bottom: 2px solid var(--color-accent-blue);
        }
        .view-panel {
            display: none;
        }
        .view-panel.active {
            display: block;
        }
    `;
    document.head.appendChild(style);

    markdownContent.querySelectorAll('.view-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            markdownContent.querySelectorAll('.view-tab').forEach(t => t.classList.remove('active'));
            markdownContent.querySelectorAll('.view-panel').forEach(p => p.classList.remove('active'));

            tab.classList.add('active');
            const viewId = tab.dataset.view + '-view';
            document.getElementById(viewId).classList.add('active');
        });
    });

    markdownContent.querySelectorAll('pre code').forEach((block) => {
        hljs.highlightElement(block);
    });
    markdownContent.querySelectorAll('.line-number').forEach(lineNum => {
        lineNum.addEventListener('click', () => {
            const line = parseInt(lineNum.dataset.line);
            showCommentForm(line);
        });
    });
}

async function loadComments() {
    try {
        const response = await fetch(`/api/markdown/${markdownId}/comments`);
        if (!response.ok) return;

        const data = await response.json();
        comments = data.comments || [];

        renderComments();
        highlightCommentedLines();
    } catch (error) {
        console.error('Error loading comments:', error);
    }
}

function renderComments() {
    commentCount.textContent = comments.length;

    if (comments.length === 0) {
        commentsList.innerHTML = '<div class="comments-empty">Click on a line number in Source view to add a comment</div>';
        return;
    }

    commentsList.innerHTML = comments.map(comment => `
        <div class="comment-item" data-line="${comment.line}">
            <div class="comment-meta">
                <span class="comment-line" data-line="${comment.line}">Line ${comment.line}</span>
                <span>${comment.author} • ${formatRelativeTime(new Date(comment.created_at))}</span>
            </div>
            <div class="comment-text">${escapeHtml(comment.text)}</div>
        </div>
    `).join('');

    commentsList.querySelectorAll('.comment-line').forEach(el => {
        el.addEventListener('click', () => {
            const line = parseInt(el.dataset.line);
            scrollToLine(line);
        });
    });
}

function highlightCommentedLines() {
    markdownContent.querySelectorAll('.code-line.has-comment').forEach(el => {
        el.classList.remove('has-comment');
    });
    const commentedLines = new Set(comments.map(c => c.line));
    commentedLines.forEach(line => {
        const lineEl = markdownContent.querySelector(`.code-line[data-line="${line}"]`);
        if (lineEl) {
            lineEl.classList.add('has-comment');
        }
    });
}

function showCommentForm(line) {
    selectedLine = line;
    commentLineNum.textContent = line;
    commentForm.classList.remove('hidden');
    commentText.focus();

    scrollToLine(line);
}

function scrollToLine(line) {
    const sourceTab = markdownContent.querySelector('[data-view="source"]');
    if (sourceTab && !sourceTab.classList.contains('active')) {
        sourceTab.click();
    }

    const lineEl = markdownContent.querySelector(`.code-line[data-line="${line}"]`);
    if (lineEl) {
        lineEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
        lineEl.style.backgroundColor = 'rgba(88, 166, 255, 0.2)';
        setTimeout(() => {
            lineEl.style.backgroundColor = '';
        }, 2000);
    }
}

function cancelComment() {
    selectedLine = null;
    commentForm.classList.add('hidden');
    commentAuthor.value = '';
    commentText.value = '';
}

async function submitComment() {
    const text = commentText.value.trim();
    if (!text) {
        showToast('Please enter a comment', 'error');
        return;
    }

    submitCommentBtn.disabled = true;
    submitCommentBtn.textContent = 'Submitting...';

    try {
        const response = await fetch(`/api/markdown/${markdownId}/comments`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                line: selectedLine,
                text: text,
                author: commentAuthor.value.trim() || 'Anonymous'
            })
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to add comment');
        }

        await loadComments();
        cancelComment();
        showToast('Comment added!', 'success');

    } catch (error) {
        showToast(error.message, 'error');
    } finally {
        submitCommentBtn.disabled = false;
        submitCommentBtn.textContent = 'Submit';
    }
}

async function deleteMarkdown() {
    if (!confirm('Are you sure you want to delete this markdown? This cannot be undone.')) {
        return;
    }

    try {
        const response = await fetch(`/api/markdown/${markdownId}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to delete');
        }

        showToast('Markdown deleted', 'success');
        setTimeout(() => {
            window.location.href = '/';
        }, 1000);

    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function copyCurrentLink() {
    try {
        await navigator.clipboard.writeText(window.location.href);
        copyLinkBtn.textContent = 'Copied!';
        setTimeout(() => {
            copyLinkBtn.textContent = 'Copy Link';
        }, 2000);
    } catch (error) {
        showToast('Failed to copy link', 'error');
    }
}

function downloadMarkdown() {
    if (!rawContent) {
        showToast('No content to download', 'error');
        return;
    }

    const firstLine = rawContent.split('\n')[0].replace(/^#*\s*/, '').trim();
    const filename = (firstLine.substring(0, 50).replace(/[^a-z0-9]/gi, '_') || 'markdown') + '.md';

    const blob = new Blob([rawContent], { type: 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    showToast('Download started!', 'success');
}

function showNotFound() {
    loadingState.classList.add('hidden');
    contentState.classList.add('hidden');
    notFoundState.classList.remove('hidden');
}

function formatRelativeTime(date) {
    const now = new Date();
    const diff = date - now;

    if (diff < 0) {
        return 'expired';
    }

    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    if (hours > 0) {
        return `in ${hours}h ${minutes}m`;
    }
    return `in ${minutes}m`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
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
