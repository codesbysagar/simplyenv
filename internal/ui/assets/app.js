// simplyenv Web UI Controller
document.addEventListener('DOMContentLoaded', () => {
  // State
  let projects = [];
  let currentProject = null;
  let currentEnvVars = {};
  let isMasked = true;
  let currentExportFormat = 'env';

  // DOM Elements
  const projectListEl = document.getElementById('projectList');
  const currentProjectNameEl = document.getElementById('currentProjectName');
  const currentProjectPathEl = document.getElementById('currentProjectPath');
  const envTableBodyEl = document.getElementById('envTableBody');
  const emptyStateEl = document.getElementById('emptyState');
  const varCountBadgeEl = document.getElementById('varCountBadge');
  const configFileBadgeEl = document.getElementById('configFileBadge');
  const searchEnvInput = document.getElementById('searchEnvInput');
  const toggleMaskAllBtn = document.getElementById('toggleMaskAllBtn');
  const maskToggleText = document.getElementById('maskToggleText');
  const toastContainer = document.getElementById('toastContainer');

  // Modals & Forms
  const varModal = document.getElementById('varModal');
  const varModalTitle = document.getElementById('varModalTitle');
  const varForm = document.getElementById('varForm');
  const varKeyInput = document.getElementById('varKeyInput');
  const varValInput = document.getElementById('varValInput');

  const projectModal = document.getElementById('projectModal');
  const projectForm = document.getElementById('projectForm');
  const projectPathInput = document.getElementById('projectPathInput');
  const useCwdBtn = document.getElementById('useCwdBtn');

  const importModal = document.getElementById('importModal');
  const importTextInput = document.getElementById('importTextInput');
  const overwriteKeysCheckbox = document.getElementById('overwriteKeysCheckbox');
  const confirmImportBtn = document.getElementById('confirmImportBtn');
  const dropZone = document.getElementById('dropZone');
  const fileInput = document.getElementById('fileInput');
  const browseFileBtn = document.getElementById('browseFileBtn');

  const exportModal = document.getElementById('exportModal');
  const exportPreviewText = document.getElementById('exportPreviewText');
  const copyExportBtn = document.getElementById('copyExportBtn');
  const downloadExportBtn = document.getElementById('downloadExportBtn');

  const shellModal = document.getElementById('shellModal');
  const shellSnippetCode = document.getElementById('shellSnippetCode');
  const shellInstructionsText = document.getElementById('shellInstructionsText');
  const copyShellHookBtn = document.getElementById('copyShellHookBtn');

  // Theme Toggle DOM
  const themeToggleBtn = document.getElementById('themeToggleBtn');
  const themeMoonIcon = document.getElementById('themeMoonIcon');
  const themeSunIcon = document.getElementById('themeSunIcon');

  // Toast Helper
  function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `<span>${message}</span>`;
    toastContainer.appendChild(toast);
    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transform = 'translateY(10px)';
      setTimeout(() => toast.remove(), 250);
    }, 2800);
  }

  // API Client
  async function api(endpoint, options = {}) {
    try {
      const res = await fetch(endpoint, {
        headers: { 'Content-Type': 'application/json', ...options.headers },
        ...options
      });
      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || res.statusText);
      }
      return await res.json();
    } catch (err) {
      showToast(err.message, 'error');
      throw err;
    }
  }

  // Load Projects
  async function loadProjects() {
    try {
      const data = await api('/api/projects');
      projects = data.projects || [];
      renderProjects();

      // If no project active yet, pick first or prompt
      if (!currentProject && projects.length > 0) {
        selectProject(projects[0]);
      } else if (projects.length === 0) {
        // Auto-detect current working directory
        detectCwd();
      }
    } catch (err) {
      console.error('Failed to load projects:', err);
    }
  }

  async function detectCwd() {
    try {
      const data = await api('/api/system/cwd');
      if (data && data.path) {
        await api('/api/projects', {
          method: 'POST',
          body: JSON.stringify({ path: data.path })
        });
        loadProjects();
      }
    } catch (e) {
      console.error('Failed to detect cwd:', e);
    }
  }

  function renderProjects() {
    projectListEl.innerHTML = '';
    projects.forEach(p => {
      const item = document.createElement('div');
      item.className = `project-item ${currentProject && currentProject.path === p.path ? 'active' : ''}`;
      item.onclick = () => selectProject(p);

      const dotClass = p.has_config ? '' : 'no-config';
      item.innerHTML = `
        <div class="project-item-info">
          <span class="project-item-name">${escapeHtml(p.name)}</span>
          <span class="project-item-path" title="${escapeHtml(p.path)}">${escapeHtml(p.path)}</span>
        </div>
        <div class="project-status-dot ${dotClass}" title="${p.has_config ? 'Configured (' + p.config_file + ')' : 'No env file yet'}"></div>
      `;
      projectListEl.appendChild(item);
    });
  }

  async function selectProject(project) {
    currentProject = project;
    currentProjectNameEl.textContent = project.name;
    currentProjectPathEl.textContent = project.path;
    configFileBadgeEl.textContent = project.config_file || '.simplyenv';
    renderProjects();
    await loadEnvironment();
  }

  async function loadEnvironment() {
    if (!currentProject) return;
    try {
      const data = await api(`/api/env?path=${encodeURIComponent(currentProject.path)}`);
      currentEnvVars = data.vars || {};
      configFileBadgeEl.textContent = data.config_file || '.simplyenv';
      renderEnvTable();
    } catch (err) {
      console.error('Failed to load environment:', err);
    }
  }

  function renderEnvTable(filter = '') {
    const query = filter.toLowerCase().trim();
    const keys = Object.keys(currentEnvVars).sort();
    const filteredKeys = keys.filter(k => k.toLowerCase().includes(query) || currentEnvVars[k].toLowerCase().includes(query));

    varCountBadgeEl.textContent = `${keys.length} ${keys.length === 1 ? 'variable' : 'variables'}`;
    envTableBodyEl.innerHTML = '';

    if (keys.length === 0) {
      emptyStateEl.classList.add('active');
    } else {
      emptyStateEl.classList.remove('active');
    }

    filteredKeys.forEach(k => {
      const val = currentEnvVars[k];
      const tr = document.createElement('tr');

      const maskedVal = isMasked ? '••••••••••••••••' : escapeHtml(val);
      const valClass = isMasked ? 'val-content masked' : 'val-content';

      tr.innerHTML = `
        <td class="key-cell">${escapeHtml(k)}</td>
        <td>
          <div class="val-cell">
            <span class="${valClass}" id="val-${k}">${maskedVal}</span>
            <button class="btn-icon" data-copy="${escapeHtml(val)}" title="Copy Value">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
              </svg>
            </button>
          </div>
        </td>
        <td style="text-align: right;">
          <div class="actions-cell">
            <button class="btn-icon" data-edit="${escapeHtml(k)}" title="Edit Variable">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
              </svg>
            </button>
            <button class="btn-icon delete" data-delete="${escapeHtml(k)}" title="Delete Variable">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
            </button>
          </div>
        </td>
      `;
      envTableBodyEl.appendChild(tr);
    });

    // Attach row events
    envTableBodyEl.querySelectorAll('[data-copy]').forEach(btn => {
      btn.onclick = () => {
        copyToClipboard(btn.getAttribute('data-copy'));
        showToast('Value copied to clipboard!');
      };
    });

    envTableBodyEl.querySelectorAll('[data-edit]').forEach(btn => {
      btn.onclick = () => openEditVarModal(btn.getAttribute('data-edit'));
    });

    envTableBodyEl.querySelectorAll('[data-delete]').forEach(btn => {
      btn.onclick = () => deleteVar(btn.getAttribute('data-delete'));
    });
  }

  // Search filter
  searchEnvInput.addEventListener('input', (e) => {
    renderEnvTable(e.target.value);
  });

  // Mask / Reveal toggle
  toggleMaskAllBtn.addEventListener('click', () => {
    isMasked = !isMasked;
    maskToggleText.textContent = isMasked ? 'Reveal Secrets' : 'Hide Secrets';
    renderEnvTable(searchEnvInput.value);
  });

  // Modal Open / Close handlers
  document.querySelectorAll('[data-close]').forEach(btn => {
    btn.onclick = () => {
      const modalId = btn.getAttribute('data-close');
      document.getElementById(modalId).classList.remove('active');
    };
  });

  // Add Variable Modal
  document.getElementById('addVarBtn').onclick = () => openAddVarModal();
  document.getElementById('emptyAddBtn').onclick = () => openAddVarModal();

  function openAddVarModal() {
    varModalTitle.textContent = 'Add Environment Variable';
    varKeyInput.value = '';
    varKeyInput.readOnly = false;
    varValInput.value = '';
    varModal.classList.add('active');
    varKeyInput.focus();
  }

  function openEditVarModal(key) {
    varModalTitle.textContent = `Edit ${key}`;
    varKeyInput.value = key;
    varKeyInput.readOnly = true;
    varValInput.value = currentEnvVars[key] || '';
    varModal.classList.add('active');
    varValInput.focus();
  }

  varForm.onsubmit = async (e) => {
    e.preventDefault();
    if (!currentProject) {
      showToast('Please select a project directory first', 'error');
      return;
    }
    const key = varKeyInput.value.trim();
    const value = varValInput.value;

    try {
      await api('/api/env', {
        method: 'POST',
        body: JSON.stringify({
          path: currentProject.path,
          key: key,
          value: value
        })
      });
      showToast(`Saved ${key}`);
      varModal.classList.remove('active');
      await loadEnvironment();
      await loadProjects();
    } catch (err) {
      console.error(err);
    }
  };

  async function deleteVar(key) {
    if (!confirm(`Are you sure you want to remove ${key}?`)) return;
    try {
      await api('/api/env', {
        method: 'DELETE',
        body: JSON.stringify({
          path: currentProject.path,
          key: key
        })
      });
      showToast(`Removed ${key}`);
      await loadEnvironment();
    } catch (err) {
      console.error(err);
    }
  }

  // Add Project Modal
  document.getElementById('addProjectBtn').onclick = () => {
    projectPathInput.value = '';
    projectModal.classList.add('active');
    projectPathInput.focus();
  };

  useCwdBtn.onclick = async () => {
    try {
      const data = await api('/api/system/cwd');
      if (data && data.path) {
        projectPathInput.value = data.path;
      }
    } catch (err) {
      console.error(err);
    }
  };

  projectForm.onsubmit = async (e) => {
    e.preventDefault();
    const pPath = projectPathInput.value.trim();
    if (!pPath) return;

    try {
      const newProj = await api('/api/projects', {
        method: 'POST',
        body: JSON.stringify({ path: pPath })
      });
      showToast('Directory tracked successfully!');
      projectModal.classList.remove('active');
      await loadProjects();
      selectProject(newProj);
    } catch (err) {
      console.error(err);
    }
  };

  // Bulk Import Modal
  document.getElementById('importBtn').onclick = () => openImportModal();
  document.getElementById('emptyImportBtn').onclick = () => openImportModal();

  function openImportModal() {
    importTextInput.value = '';
    importModal.classList.add('active');
  }

  // Import Tabs
  document.querySelectorAll('[data-import-tab]').forEach(tabBtn => {
    tabBtn.onclick = () => {
      document.querySelectorAll('[data-import-tab]').forEach(b => b.classList.remove('active'));
      tabBtn.classList.add('active');
      const tab = tabBtn.getAttribute('data-import-tab');
      if (tab === 'paste') {
        document.getElementById('pasteTabContent').classList.remove('hidden');
        document.getElementById('fileTabContent').classList.add('hidden');
      } else {
        document.getElementById('pasteTabContent').classList.add('hidden');
        document.getElementById('fileTabContent').classList.remove('hidden');
      }
    };
  });

  // Drag and drop & file upload
  browseFileBtn.onclick = () => fileInput.click();
  dropZone.onclick = (e) => {
    if (e.target !== browseFileBtn) fileInput.click();
  };

  fileInput.onchange = (e) => {
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0]);
    }
  };

  dropZone.ondragover = (e) => {
    e.preventDefault();
    dropZone.classList.add('dragover');
  };

  dropZone.ondragleave = () => {
    dropZone.classList.remove('dragover');
  };

  dropZone.ondrop = (e) => {
    e.preventDefault();
    dropZone.classList.remove('dragover');
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0]);
    }
  };

  function handleFile(file) {
    const reader = new FileReader();
    reader.onload = (event) => {
      importTextInput.value = event.target.result;
      document.querySelector('[data-import-tab="paste"]').click();
      showToast(`Loaded ${file.name}`);
    };
    reader.readAsText(file);
  }

  confirmImportBtn.onclick = async () => {
    if (!currentProject) {
      showToast('Select a directory first', 'error');
      return;
    }
    const content = importTextInput.value.trim();
    if (!content) {
      showToast('Please provide env content to import', 'error');
      return;
    }

    try {
      const res = await api('/api/import', {
        method: 'POST',
        body: JSON.stringify({
          path: currentProject.path,
          content: content,
          overwrite: overwriteKeysCheckbox.checked
        })
      });
      showToast(`Successfully imported ${res.count} variables!`);
      importModal.classList.remove('active');
      await loadEnvironment();
      await loadProjects();
    } catch (err) {
      console.error(err);
    }
  };

  // Bulk Export Modal
  document.getElementById('exportBtn').onclick = () => openExportModal();

  async function openExportModal() {
    if (!currentProject) {
      showToast('Select a directory first', 'error');
      return;
    }
    exportModal.classList.add('active');
    await updateExportPreview();
  }

  document.querySelectorAll('[data-format]').forEach(tabBtn => {
    tabBtn.onclick = async () => {
      document.querySelectorAll('[data-format]').forEach(b => b.classList.remove('active'));
      tabBtn.classList.add('active');
      currentExportFormat = tabBtn.getAttribute('data-format');
      await updateExportPreview();
    };
  });

  async function updateExportPreview() {
    try {
      const res = await api(`/api/export?path=${encodeURIComponent(currentProject.path)}&format=${currentExportFormat}`);
      exportPreviewText.value = res.content || '';
    } catch (err) {
      console.error(err);
    }
  }

  copyExportBtn.onclick = () => {
    copyToClipboard(exportPreviewText.value);
    showToast('Export content copied to clipboard!');
  };

  downloadExportBtn.onclick = () => {
    const blob = new Blob([exportPreviewText.value], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    let ext = '.env';
    if (currentExportFormat === 'json') ext = '.json';
    if (currentExportFormat === 'shell') ext = '.sh';
    if (currentExportFormat === 'docker') ext = '.dockerfile';
    a.download = `${currentProject.name}_env${ext}`;
    a.click();
    URL.revokeObjectURL(url);
    showToast('File downloaded!');
  };

  // Shell Setup Modal & Auto-Install
  let activeShellTab = 'bash';
  let isHookInstalledCurrently = false;

  const shellInstallBadge = document.getElementById('shellInstallBadge');
  const shellProfilePathText = document.getElementById('shellProfilePathText');
  const autoInstallHookBtn = document.getElementById('autoInstallHookBtn');
  const autoInstallBtnText = document.getElementById('autoInstallBtnText');

  document.getElementById('openShellModalBtn').onclick = async () => {
    shellModal.classList.add('active');
    await checkShellStatus('bash');
  };

  document.querySelectorAll('[data-shell]').forEach(btn => {
    btn.onclick = async () => {
      document.querySelectorAll('[data-shell]').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      activeShellTab = btn.getAttribute('data-shell');
      await checkShellStatus(activeShellTab);
    };
  });

  async function checkShellStatus(shellName) {
    updateShellSnippet(shellName);
    try {
      const data = await api(`/api/system/shell?shell=${encodeURIComponent(shellName)}`);
      shellProfilePathText.textContent = data.profile_path || 'Unknown profile';
      isHookInstalledCurrently = data.installed;

      if (data.installed) {
        shellInstallBadge.className = 'badge badge-success';
        shellInstallBadge.textContent = 'Installed';
        autoInstallBtnText.textContent = 'Remove from Profile';
        autoInstallHookBtn.className = 'btn btn-secondary';
      } else {
        shellInstallBadge.className = 'badge badge-warning';
        shellInstallBadge.textContent = 'Not Installed';
        autoInstallBtnText.textContent = `Install into ${shellName.toUpperCase()}`;
        autoInstallHookBtn.className = 'btn btn-primary';
      }
    } catch (err) {
      console.error('Failed to check shell status:', err);
    }
  }

  autoInstallHookBtn.onclick = async () => {
    try {
      if (isHookInstalledCurrently) {
        const res = await api('/api/system/uninstall-hook', {
          method: 'POST',
          body: JSON.stringify({ shell: activeShellTab })
        });
        showToast(res.message || 'Hook removed successfully');
      } else {
        const res = await api('/api/system/install-hook', {
          method: 'POST',
          body: JSON.stringify({ shell: activeShellTab })
        });
        showToast(res.message || 'Hook installed automatically!');
      }
      await checkShellStatus(activeShellTab);
    } catch (err) {
      console.error(err);
    }
  };

  function updateShellSnippet(shell) {
    switch (shell) {
      case 'zsh':
        shellSnippetCode.textContent = 'eval "$(simplyenv hook zsh)"';
        shellInstructionsText.innerHTML = 'Or manually add the line above to your <code>~/.zshrc</code> file.';
        break;
      case 'fish':
        shellSnippetCode.textContent = 'simplyenv hook fish | source';
        shellInstructionsText.innerHTML = 'Or manually add the line above to your <code>~/.config/fish/config.fish</code> file.';
        break;
      case 'pwsh':
        shellSnippetCode.textContent = 'Invoke-Expression (simplyenv hook pwsh)';
        shellInstructionsText.innerHTML = 'Or manually add the line above to your <code>$PROFILE</code> file in PowerShell.';
        break;
      case 'bash':
      default:
        shellSnippetCode.textContent = 'eval "$(simplyenv hook bash)"';
        shellInstructionsText.innerHTML = 'Or manually add the line above to your <code>~/.bashrc</code> file.';
        break;
    }
  }

  copyShellHookBtn.onclick = () => {
    copyToClipboard(shellSnippetCode.textContent);
    showToast('Shell hook copied!');
  };

  // Theme Management (Light by default)
  function initTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    updateThemeIcons(currentTheme);
  }

  function updateThemeIcons(theme) {
    if (theme === 'dark') {
      themeMoonIcon.classList.add('hidden');
      themeSunIcon.classList.remove('hidden');
    } else {
      themeMoonIcon.classList.remove('hidden');
      themeSunIcon.classList.add('hidden');
    }
  }

  if (themeToggleBtn) {
    themeToggleBtn.onclick = () => {
      const current = document.documentElement.getAttribute('data-theme') || 'light';
      const next = current === 'dark' ? 'light' : 'dark';
      document.documentElement.setAttribute('data-theme', next);
      localStorage.setItem('simplyenv_theme', next);
      updateThemeIcons(next);
      showToast(`Switched to ${next} mode`);
    };
  }

  // Utilities
  function copyToClipboard(text) {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(text);
    } else {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      textarea.remove();
    }
  }

  function escapeHtml(str) {
    if (typeof str !== 'string') return '';
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  // Interactive Tour Guide
  let currentTourStep = 0;
  const tourModal = document.getElementById('tourModal');
  const tourBody = document.getElementById('tourBody');
  const tourStepBadge = document.getElementById('tourStepBadge');
  const tourDots = document.getElementById('tourDots');
  const tourPrevBtn = document.getElementById('tourPrevBtn');
  const tourNextBtn = document.getElementById('tourNextBtn');
  const openTourBtn = document.getElementById('openTourBtn');

  const tourSteps = [
    {
      title: "Welcome to simplyenv",
      icon: `<img src="/logo-icon.png" alt="simplyenv" class="tour-logo-icon">`,
      desc: "simplyenv solves the pain of managing directory-specific environments. It pairs automatic shell directory switching with a modern, visual dashboard and developer-friendly tooling.",
      features: [
        "Automatic environment loading as you cd into project folders",
        "Interactive UI to inspect, modify, and manage variables visually",
        "100% standalone single binary with zero external dependencies"
      ]
    },
    {
      title: "1. Projects Catalog (Sidebar)",
      icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>`,
      desc: "The left sidebar acts as your central command center. You can view all tracked repositories and directories without having to navigate to them in your terminal.",
      features: [
        "Green indicator shows whether a .simplyenv configuration exists",
        "Click the '+' button to track any directory on your computer",
        "Switch projects instantly to inspect and configure their variables"
      ]
    },
    {
      title: "2. Variable Table & Secret Masking",
      icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>`,
      desc: "Your environment keys and values are displayed in a clean, monospace table designed for readability and security.",
      features: [
        "Values are masked by default so you can safely share your screen",
        "Toggle 'Reveal Secrets' to view API tokens or database passwords",
        "One-click copy buttons next to every value and variable name",
        "Search in real time using the quick filter search box"
      ]
    },
    {
      title: "3. Bulk Import & Multi-Format Export",
      icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>`,
      desc: "Easily migrate existing configurations and share environments across your deployment pipelines.",
      features: [
        "Import: Drag and drop any .env, JSON, or key-value file with live preview",
        "Export: Convert your variables to .env, JSON, Shell export, or Dockerfile ENV",
        "Download as a file or copy directly to your clipboard in one click"
      ]
    },
    {
      title: "4. 1-Click Shell Integration",
      icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line></svg>`,
      desc: "Forget manual copy-pasting into dotfiles! simplyenv can configure your terminal shell automatically.",
      features: [
        "Click 'Shell Hook Setup' in the sidebar footer",
        "Press 'Install Automatically' to configure ~/.bashrc, ~/.zshrc, or Fish",
        "Now whenever you cd into a directory, variables load effortlessly!"
      ]
    }
  ];

  function openTourModal(step = 0) {
    currentTourStep = step;
    renderTour();
    tourModal.classList.add('active');
  }

  function renderTour() {
    const s = tourSteps[currentTourStep];
    tourStepBadge.textContent = `Step ${currentTourStep + 1} of ${tourSteps.length}`;

    tourBody.innerHTML = `
      <div class="tour-slide">
        <div class="tour-icon-box">${s.icon}</div>
        <h4>${s.title}</h4>
        <p>${s.desc}</p>
        <div class="tour-features-list">
          ${s.features.map(f => `
            <div class="tour-feature-item">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
              <span>${f}</span>
            </div>
          `).join('')}
        </div>
      </div>
    `;

    // Render Dots
    tourDots.innerHTML = '';
    tourSteps.forEach((_, idx) => {
      const dot = document.createElement('span');
      dot.className = `tour-dot ${idx === currentTourStep ? 'active' : ''}`;
      dot.onclick = () => {
        currentTourStep = idx;
        renderTour();
      };
      tourDots.appendChild(dot);
    });

    // Button states
    tourPrevBtn.disabled = currentTourStep === 0;
    if (currentTourStep === tourSteps.length - 1) {
      tourNextBtn.textContent = 'Finish Tour';
    } else {
      tourNextBtn.textContent = 'Next';
    }
  }

  tourPrevBtn.onclick = () => {
    if (currentTourStep > 0) {
      currentTourStep--;
      renderTour();
    }
  };

  tourNextBtn.onclick = () => {
    if (currentTourStep < tourSteps.length - 1) {
      currentTourStep++;
      renderTour();
    } else {
      tourModal.classList.remove('active');
      localStorage.setItem('simplyenv_tour_seen', 'true');
      showToast('Tour completed! Click "Tour" anytime to revisit.');
    }
  };

  if (openTourBtn) {
    openTourBtn.onclick = () => openTourModal(0);
  }

  // Initial Boot
  initTheme();
  loadProjects();

  // Auto-launch tour on first visit
  if (!localStorage.getItem('simplyenv_tour_seen')) {
    setTimeout(() => openTourModal(0), 400);
  }
});
