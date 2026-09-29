// simplyenv Web UI Controller — Project & Module Edition
document.addEventListener('DOMContentLoaded', () => {
  // State
  let projects = [];
  let currentProject = null;
  let currentModule = null;
  let currentView = 'overview'; // 'overview' | 'module'
  let expandedProjects = new Set();
  let currentEnvVars = {};
  let isMasked = true;
  let currentExportFormat = 'env';

  // DOM Elements - Sidebar & Topbar
  const projectListEl = document.getElementById('projectList');
  const breadcrumbProjectBtn = document.getElementById('breadcrumbProjectBtn');
  const breadcrumbProjectName = document.getElementById('breadcrumbProjectName');
  const breadcrumbSep = document.getElementById('breadcrumbSep');
  const moduleDropdownWrapper = document.getElementById('moduleDropdownWrapper');
  const moduleSelectTrigger = document.getElementById('moduleSelectTrigger');
  const moduleTriggerIcon = document.getElementById('moduleTriggerIcon');
  const breadcrumbModuleName = document.getElementById('breadcrumbModuleName');
  const moduleDropdownMenu = document.getElementById('moduleDropdownMenu');
  const currentPathDisplay = document.getElementById('currentPathDisplay');
  const searchEnvInput = document.getElementById('searchEnvInput');
  const toastContainer = document.getElementById('toastContainer');
  const addModuleTopBtn = document.getElementById('addModuleTopBtn');

  // DOM Elements - Views
  const projectOverviewView = document.getElementById('projectOverviewView');
  const overviewProjectName = document.getElementById('overviewProjectName');
  const overviewProjectPath = document.getElementById('overviewProjectPath');
  const overviewModuleCount = document.getElementById('overviewModuleCount');
  const overviewTotalVars = document.getElementById('overviewTotalVars');
  const overviewAddModuleBtn = document.getElementById('overviewAddModuleBtn');
  const modulesGrid = document.getElementById('modulesGrid');

  const moduleEnvView = document.getElementById('moduleEnvView');
  const backToOverviewBtn = document.getElementById('backToOverviewBtn');
  const backProjectName = document.getElementById('backProjectName');
  const activeModuleBadge = document.getElementById('activeModuleBadge');
  const varCountBadgeEl = document.getElementById('varCountBadge');
  const configFileBadgeEl = document.getElementById('configFileBadge');
  const toggleMaskAllBtn = document.getElementById('toggleMaskAllBtn');
  const maskToggleText = document.getElementById('maskToggleText');
  const envTableBodyEl = document.getElementById('envTableBody');
  const emptyStateEl = document.getElementById('emptyState');
  const emptyStateTitle = document.getElementById('emptyStateTitle');
  const emptyStateDesc = document.getElementById('emptyStateDesc');

  // Modals & Forms
  const varModal = document.getElementById('varModal');
  const varModalTitle = document.getElementById('varModalTitle');
  const varModalModulePill = document.getElementById('varModalModulePill');
  const varForm = document.getElementById('varForm');
  const varKeyInput = document.getElementById('varKeyInput');
  const varValInput = document.getElementById('varValInput');

  const moduleModal = document.getElementById('moduleModal');
  const moduleForm = document.getElementById('moduleForm');
  const moduleModalProjectName = document.getElementById('moduleModalProjectName');
  const moduleNameInput = document.getElementById('moduleNameInput');
  const moduleRelPathInput = document.getElementById('moduleRelPathInput');
  const modulePathPreview = document.getElementById('modulePathPreview');

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

  // Active target helper
  function getActiveTarget() {
    if (currentModule) {
      return {
        path: currentModule.path,
        name: currentModule.name,
        isModule: true,
        configFile: currentModule.config_file || '.simplyenv'
      };
    }
    if (currentProject) {
      return {
        path: currentProject.path,
        name: currentProject.name,
        isModule: false,
        configFile: currentProject.config_file || '.simplyenv'
      };
    }
    return null;
  }

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

      // Re-link currentProject & currentModule references if they exist
      if (currentProject) {
        const found = projects.find(p => p.path === currentProject.path);
        currentProject = found || null;
      }
      if (currentProject && currentModule) {
        const foundMod = (currentProject.modules || []).find(m => m.path === currentModule.path);
        currentModule = foundMod || null;
      }

      // If no project active yet, default to first or auto-detect cwd
      if (!currentProject && projects.length > 0) {
        const p = projects[0];
        expandedProjects.add(p.path);
        if (p.modules && p.modules.length > 0) {
          selectModule(p, p.modules[0]);
        } else {
          selectProject(p, true);
        }
      } else if (projects.length === 0) {
        detectCwd();
      } else {
        renderProjects();
        if (currentView === 'overview' && currentProject) {
          renderProjectOverview(currentProject);
        } else if (currentView === 'module' && currentModule) {
          await loadEnvironment();
        }
      }
    } catch (err) {
      console.error('Failed to load projects:', err);
    }
  }

  async function detectCwd() {
    try {
      const data = await api('/api/system/cwd');
      if (data && data.path) {
        const newProj = await api('/api/projects', {
          method: 'POST',
          body: JSON.stringify({ path: data.path })
        });
        await loadProjects();
      }
    } catch (e) {
      console.error('Failed to detect cwd:', e);
    }
  }

  // Render Projects & Modules Tree
  function renderProjects() {
    projectListEl.innerHTML = '';

    projects.forEach(p => {
      const isExpanded = expandedProjects.has(p.path);
      const isProjectActive = currentProject && currentProject.path === p.path && currentView === 'overview';
      const modules = p.modules || [];

      const groupEl = document.createElement('div');
      groupEl.className = `project-group ${isExpanded ? 'expanded' : ''}`;

      // Project Header row
      const headerEl = document.createElement('div');
      headerEl.className = `project-header ${isProjectActive ? 'active' : ''}`;

      headerEl.innerHTML = `
        <div class="project-header-left">
          <svg class="project-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="9 18 15 12 9 6"></polyline>
          </svg>
          <svg class="project-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
          </svg>
          <span class="project-title" title="${escapeHtml(p.path)}">${escapeHtml(p.name)}</span>
        </div>
        <div class="project-header-right">
          <span class="module-count-pill" title="${modules.length} microservices / modules">${modules.length}</span>
          <div class="project-quick-actions">
            <button class="btn-icon-xs add-mod-btn" title="Add Module to ${escapeHtml(p.name)}">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
            </button>
            <button class="btn-icon-xs delete del-proj-btn" title="Untrack Project">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
            </button>
          </div>
        </div>
      `;

      // Header click interactions
      headerEl.onclick = (e) => {
        if (e.target.closest('.add-mod-btn')) {
          e.stopPropagation();
          openAddModuleModal(p);
          return;
        }
        if (e.target.closest('.del-proj-btn')) {
          e.stopPropagation();
          deleteProject(p);
          return;
        }
        if (e.target.closest('.project-chevron')) {
          e.stopPropagation();
          toggleProjectExpand(p.path);
          return;
        }

        // Clicking anywhere else on project header opens Project Overview
        selectProject(p, true);
      };

      // Modules List inside project
      const moduleListEl = document.createElement('div');
      moduleListEl.className = 'module-list';

      modules.forEach(m => {
        const isModuleActive = currentProject && currentProject.path === p.path &&
                               currentModule && currentModule.path === m.path &&
                               currentView === 'module';

        const itemEl = document.createElement('div');
        itemEl.className = `module-item ${isModuleActive ? 'active' : ''}`;

        const isRoot = m.rel_path === '.';
        const iconEmoji = isRoot ? '⚡' : '📦';
        const displayName = isRoot ? (p.name + ' (root)') : m.name;
        const varCount = m.var_count || 0;
        const dotClass = m.has_config ? '' : 'no-config';

        itemEl.innerHTML = `
          <div class="module-item-left" title="${escapeHtml(m.path)}">
            <span class="module-icon-emoji">${iconEmoji}</span>
            <span class="module-name">${escapeHtml(displayName)}</span>
          </div>
          <div class="module-item-right">
            <span class="module-vars-pill">${varCount} vars</span>
            <button class="btn-icon-xs delete del-mod-btn" title="Remove Module">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
            <div class="module-status-dot ${dotClass}" title="${m.has_config ? 'Configured' : 'No config yet'}"></div>
          </div>
        `;

        itemEl.onclick = (e) => {
          if (e.target.closest('.del-mod-btn')) {
            e.stopPropagation();
            deleteModule(p, m);
            return;
          }
          selectModule(p, m);
        };

        moduleListEl.appendChild(itemEl);
      });

      // "+ Add Module" link at bottom of module list
      const addModLink = document.createElement('button');
      addModLink.className = 'add-module-sidebar-btn';
      addModLink.innerHTML = `
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        <span>Add microservice module</span>
      `;
      addModLink.onclick = (e) => {
        e.stopPropagation();
        openAddModuleModal(p);
      };
      moduleListEl.appendChild(addModLink);

      groupEl.appendChild(headerEl);
      groupEl.appendChild(moduleListEl);
      projectListEl.appendChild(groupEl);
    });
  }

  function toggleProjectExpand(projectPath) {
    if (expandedProjects.has(projectPath)) {
      expandedProjects.delete(projectPath);
    } else {
      expandedProjects.add(projectPath);
    }
    renderProjects();
  }

  // Select Project & Show Overview Dashboard
  function selectProject(project, openOverview = true) {
    currentProject = project;
    expandedProjects.add(project.path);

    if (openOverview || !project.modules || project.modules.length === 0) {
      currentView = 'overview';
      currentModule = null;
      renderProjectOverview(project);
    } else {
      selectModule(project, project.modules[0]);
    }
    renderProjects();
  }

  function renderProjectOverview(project) {
    currentView = 'overview';
    projectOverviewView.classList.remove('hidden');
    moduleEnvView.classList.add('hidden');

    // Breadcrumb updates
    breadcrumbProjectName.textContent = project.name;
    breadcrumbSep.classList.add('hidden');
    moduleDropdownWrapper.classList.add('hidden');
    currentPathDisplay.textContent = project.path;

    // Overview Hero Card
    overviewProjectName.textContent = project.name;
    overviewProjectPath.textContent = project.path;

    const modules = project.modules || [];
    overviewModuleCount.textContent = modules.length;

    let totalVars = 0;
    modules.forEach(m => { totalVars += (m.var_count || 0); });
    overviewTotalVars.textContent = totalVars;

    // Render Modules Grid
    modulesGrid.innerHTML = '';

    modules.forEach(m => {
      const isRoot = m.rel_path === '.';
      const iconEmoji = isRoot ? '⚡' : '📦';
      const displayName = isRoot ? (project.name + ' (root)') : m.name;
      const relPathText = isRoot ? './ (root)' : `./${m.rel_path}`;

      const card = document.createElement('div');
      card.className = 'module-card';
      card.innerHTML = `
        <div class="module-card-header">
          <div class="module-card-title-group">
            <span class="module-card-icon">${iconEmoji}</span>
            <div>
              <h4 class="module-card-name" title="${escapeHtml(m.name)}">${escapeHtml(displayName)}</h4>
              <span class="module-card-relpath" title="${escapeHtml(m.path)}">${escapeHtml(relPathText)}</span>
            </div>
          </div>
          <button class="btn-icon-xs delete card-del-btn" title="Remove Module">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
          </button>
        </div>
        <div class="module-card-meta">
          <span class="badge ${m.has_config ? 'badge-module' : 'badge-outline'}">${m.var_count || 0} variables</span>
          <span class="badge badge-outline">${escapeHtml(m.config_file || '.simplyenv')}</span>
        </div>
        <div class="module-card-footer">
          <button class="btn btn-primary btn-sm manage-mod-btn">Manage Environment</button>
          <button class="btn btn-secondary btn-sm export-mod-btn" title="Export this module">Export</button>
        </div>
      `;

      card.querySelector('.manage-mod-btn').onclick = () => selectModule(project, m);
      card.querySelector('.export-mod-btn').onclick = (e) => {
        e.stopPropagation();
        currentModule = m;
        openExportModal();
      };
      card.querySelector('.card-del-btn').onclick = (e) => {
        e.stopPropagation();
        deleteModule(project, m);
      };

      modulesGrid.appendChild(card);
    });

    // Add Module Dashed Card
    const addCard = document.createElement('div');
    addCard.className = 'module-card dashed-add';
    addCard.innerHTML = `
      <div class="add-card-content">
        <div class="add-card-icon">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
        </div>
        <span class="add-card-text">Add Microservice Module</span>
        <span class="add-card-sub">Group env tracking inside ${escapeHtml(project.name)}</span>
      </div>
    `;
    addCard.onclick = () => openAddModuleModal(project);
    modulesGrid.appendChild(addCard);
  }

  // Select Specific Module & Open Environment Editor
  async function selectModule(project, module) {
    currentProject = project;
    currentModule = module;
    currentView = 'module';
    expandedProjects.add(project.path);

    projectOverviewView.classList.add('hidden');
    moduleEnvView.classList.remove('hidden');

    // Breadcrumb updates
    breadcrumbProjectName.textContent = project.name;
    breadcrumbSep.classList.remove('hidden');
    moduleDropdownWrapper.classList.remove('hidden');

    const isRoot = module.rel_path === '.';
    breadcrumbModuleName.textContent = module.name;
    moduleTriggerIcon.textContent = isRoot ? '⚡' : '📦';
    currentPathDisplay.textContent = module.path;

    // View Badges
    activeModuleBadge.textContent = `${isRoot ? '⚡' : '📦'} ${module.name}`;
    configFileBadgeEl.textContent = module.config_file || '.simplyenv';
    backProjectName.textContent = project.name;

    renderModuleDropdownMenu(project);
    renderProjects();
    await loadEnvironment();
  }

  function renderModuleDropdownMenu(project) {
    moduleDropdownMenu.innerHTML = '';
    const modules = project.modules || [];

    modules.forEach(m => {
      const btn = document.createElement('button');
      btn.className = `module-dropdown-item ${currentModule && currentModule.path === m.path ? 'active' : ''}`;
      const isRoot = m.rel_path === '.';
      btn.innerHTML = `
        <span>${isRoot ? '⚡' : '📦'} ${escapeHtml(m.name)}</span>
        <span class="module-vars-pill">${m.var_count || 0} vars</span>
      `;
      btn.onclick = () => {
        moduleDropdownMenu.classList.add('hidden');
        selectModule(project, m);
      };
      moduleDropdownMenu.appendChild(btn);
    });

    const addBtn = document.createElement('button');
    addBtn.className = 'module-dropdown-item';
    addBtn.style.color = 'var(--primary)';
    addBtn.style.fontWeight = '600';
    addBtn.innerHTML = `<span>+ Add new module</span>`;
    addBtn.onclick = () => {
      moduleDropdownMenu.classList.add('hidden');
      openAddModuleModal(project);
    };
    moduleDropdownMenu.appendChild(addBtn);
  }

  // Toggle Module Dropdown
  moduleSelectTrigger.onclick = (e) => {
    e.stopPropagation();
    moduleDropdownMenu.classList.toggle('hidden');
  };

  document.addEventListener('click', (e) => {
    if (!moduleDropdownWrapper.contains(e.target)) {
      moduleDropdownMenu.classList.add('hidden');
    }
  });

  breadcrumbProjectBtn.onclick = () => {
    if (currentProject) {
      selectProject(currentProject, true);
    }
  };

  backToOverviewBtn.onclick = () => {
    if (currentProject) {
      selectProject(currentProject, true);
    }
  };

  // Load Environment Variables for Active Module or Project
  async function loadEnvironment() {
    const target = getActiveTarget();
    if (!target) return;

    try {
      const data = await api(`/api/env?path=${encodeURIComponent(target.path)}`);
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

    const target = getActiveTarget();
    if (target) {
      emptyStateTitle.textContent = `No Variables in ${target.name}`;
      emptyStateDesc.textContent = `Add environment variables for ${target.name}, or bulk import from an existing .env file.`;
    }

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

    // Row event handlers
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

  // Add / Edit Variable Modal
  document.getElementById('addVarBtn').onclick = () => openAddVarModal();
  document.getElementById('emptyAddBtn').onclick = () => openAddVarModal();

  function openAddVarModal() {
    const target = getActiveTarget();
    if (!target) {
      showToast('Please select a project or module first', 'error');
      return;
    }
    varModalTitle.textContent = `Add Variable to ${target.name}`;
    varModalModulePill.textContent = target.name;
    varKeyInput.value = '';
    varKeyInput.readOnly = false;
    varValInput.value = '';
    varModal.classList.add('active');
    varKeyInput.focus();
  }

  function openEditVarModal(key) {
    const target = getActiveTarget();
    varModalTitle.textContent = `Edit ${key}`;
    varModalModulePill.textContent = target ? target.name : 'Active';
    varKeyInput.value = key;
    varKeyInput.readOnly = true;
    varValInput.value = currentEnvVars[key] || '';
    varModal.classList.add('active');
    varValInput.focus();
  }

  varForm.onsubmit = async (e) => {
    e.preventDefault();
    const target = getActiveTarget();
    if (!target) {
      showToast('Please select a module first', 'error');
      return;
    }
    const key = varKeyInput.value.trim();
    const value = varValInput.value;

    try {
      await api('/api/env', {
        method: 'POST',
        body: JSON.stringify({
          path: target.path,
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
    const target = getActiveTarget();
    if (!target) return;
    if (!confirm(`Are you sure you want to remove ${key}?`)) return;

    try {
      await api('/api/env', {
        method: 'DELETE',
        body: JSON.stringify({
          path: target.path,
          key: key
        })
      });
      showToast(`Removed ${key}`);
      await loadEnvironment();
      await loadProjects();
    } catch (err) {
      console.error(err);
    }
  }

  // ==========================================
  // Add Microservice / Module Modal Handlers
  // ==========================================
  function openAddModuleModal(project) {
    const proj = project || currentProject;
    if (!proj) {
      showToast('Select or create a project first', 'error');
      return;
    }
    moduleModalProjectName.textContent = proj.name;
    moduleNameInput.value = '';
    moduleRelPathInput.value = '';
    updateModulePathPreview(proj);
    moduleModal.classList.add('active');
    moduleNameInput.focus();
  }

  function updateModulePathPreview(proj) {
    const p = proj || currentProject;
    if (!p) return;
    const name = moduleNameInput.value.trim();
    const rel = moduleRelPathInput.value.trim() || name || 'microservice-name';
    modulePathPreview.textContent = `${p.path}/${rel}/.simplyenv`;
  }

  moduleNameInput.addEventListener('input', () => {
    if (!moduleRelPathInput.dataset.userEdited) {
      moduleRelPathInput.value = moduleNameInput.value.toLowerCase().replace(/[^a-z0-9-_]/g, '-');
    }
    updateModulePathPreview();
  });

  moduleRelPathInput.addEventListener('input', () => {
    moduleRelPathInput.dataset.userEdited = 'true';
    updateModulePathPreview();
  });

  // Preset Chips
  document.querySelectorAll('.preset-chip').forEach(chip => {
    chip.onclick = () => {
      const preset = chip.getAttribute('data-preset');
      moduleNameInput.value = preset;
      moduleRelPathInput.value = preset;
      delete moduleRelPathInput.dataset.userEdited;
      updateModulePathPreview();
      moduleNameInput.focus();
    };
  });

  moduleForm.onsubmit = async (e) => {
    e.preventDefault();
    if (!currentProject) {
      showToast('No active project', 'error');
      return;
    }
    const name = moduleNameInput.value.trim();
    const relPath = moduleRelPathInput.value.trim() || name;

    try {
      const newMod = await api('/api/modules', {
        method: 'POST',
        body: JSON.stringify({
          project_path: currentProject.path,
          name: name,
          rel_path: relPath
        })
      });
      showToast(`Microservice module "${name}" created!`);
      moduleModal.classList.remove('active');
      await loadProjects();
      selectModule(currentProject, newMod);
    } catch (err) {
      console.error(err);
    }
  };

  overviewAddModuleBtn.onclick = () => openAddModuleModal(currentProject);
  addModuleTopBtn.onclick = () => openAddModuleModal(currentProject);

  async function deleteModule(project, module) {
    if (!confirm(`Are you sure you want to stop tracking microservice "${module.name}"?`)) return;

    try {
      await api('/api/modules', {
        method: 'DELETE',
        body: JSON.stringify({
          project_path: project.path,
          module_path: module.path
        })
      });
      showToast(`Module "${module.name}" removed from project`);
      await loadProjects();
      selectProject(project, true);
    } catch (err) {
      console.error(err);
    }
  }

  // ==========================================
  // Track Project Modal Handlers
  // ==========================================
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
      showToast('Project tracked successfully!');
      projectModal.classList.remove('active');
      await loadProjects();
      selectProject(newProj, true);
    } catch (err) {
      console.error(err);
    }
  };

  async function deleteProject(project) {
    if (!confirm(`Are you sure you want to untrack project "${project.name}"?`)) return;

    try {
      await api('/api/projects', {
        method: 'DELETE',
        body: JSON.stringify({ path: project.path })
      });
      showToast(`Untracked project "${project.name}"`);
      currentProject = null;
      currentModule = null;
      await loadProjects();
    } catch (err) {
      console.error(err);
    }
  }

  // ==========================================
  // Bulk Import Modal
  // ==========================================
  document.getElementById('importBtn').onclick = () => openImportModal();
  document.getElementById('emptyImportBtn').onclick = () => openImportModal();

  function openImportModal() {
    const target = getActiveTarget();
    if (!target) {
      showToast('Select a module or project first', 'error');
      return;
    }
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
    const target = getActiveTarget();
    if (!target) {
      showToast('Select a module or project first', 'error');
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
          path: target.path,
          content: content,
          overwrite: overwriteKeysCheckbox.checked
        })
      });
      showToast(`Successfully imported ${res.count} variables into ${target.name}!`);
      importModal.classList.remove('active');
      await loadEnvironment();
      await loadProjects();
    } catch (err) {
      console.error(err);
    }
  };

  // ==========================================
  // Bulk Export Modal
  // ==========================================
  document.getElementById('exportBtn').onclick = () => openExportModal();

  async function openExportModal() {
    const target = getActiveTarget();
    if (!target) {
      showToast('Select a module or project first', 'error');
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
    const target = getActiveTarget();
    if (!target) return;
    try {
      const res = await api(`/api/export?path=${encodeURIComponent(target.path)}&format=${currentExportFormat}`);
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
    const target = getActiveTarget();
    const blob = new Blob([exportPreviewText.value], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    let ext = '.env';
    if (currentExportFormat === 'json') ext = '.json';
    if (currentExportFormat === 'shell') ext = '.sh';
    if (currentExportFormat === 'docker') ext = '.dockerfile';
    const targetName = target ? target.name.replace(/[^a-zA-Z0-9_-]/g, '_') : 'env';
    a.download = `${targetName}_env${ext}`;
    a.click();
    URL.revokeObjectURL(url);
    showToast('File downloaded!');
  };

  // ==========================================
  // Shell Setup Modal & Auto-Install
  // ==========================================
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

  // ==========================================
  // Theme Management
  // ==========================================
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

  // ==========================================
  // Interactive Tour Guide
  // ==========================================
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
      desc: "simplyenv solves the pain of managing directory-specific environments. It pairs automatic shell directory switching with a modern, visual dashboard and microservice-ready project grouping.",
      features: [
        "Automatic environment loading as you cd into project folders",
        "Project & Microservice Module grouping for monorepos & microservices",
        "Interactive UI to inspect, modify, and manage variables visually",
        "100% standalone single binary with zero external dependencies"
      ]
    },
    {
      title: "1. Projects & Microservices Catalog",
      icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>`,
      desc: "Organize projects with multiple services cleanly! Group microservice-A, microservice-B, or frontend/backend under a single parent project.",
      features: [
        "Expandable project accordion displays all nested microservices",
        "Automatic discovery of any subdirectories with .simplyenv",
        "Click '+ Module' to create a new microservice environment in seconds"
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
        "Now whenever you cd into a microservice directory, its variables load effortlessly!"
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

  // Sidebar Collapse / Expand Toggle
  const appLayoutEl = document.querySelector('.app-layout');
  const sidebarToggleBtn = document.getElementById('sidebarToggleBtn');

  function toggleSidebar(forceState) {
    if (!appLayoutEl) return;
    const isCurrentlyCollapsed = appLayoutEl.classList.contains('sidebar-collapsed');
    const newState = typeof forceState === 'boolean' ? forceState : !isCurrentlyCollapsed;
    appLayoutEl.classList.toggle('sidebar-collapsed', newState);
    localStorage.setItem('simplyenv_sidebar_collapsed', newState ? 'true' : 'false');
  }

  if (sidebarToggleBtn) {
    sidebarToggleBtn.onclick = () => toggleSidebar();
  }

  // Restore sidebar state
  if (localStorage.getItem('simplyenv_sidebar_collapsed') === 'true') {
    appLayoutEl?.classList.add('sidebar-collapsed');
  }

  // Keyboard shortcut: Ctrl+F / Cmd+F to focus search filter, Ctrl+B / Cmd+B for sidebar
  window.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') {
      if (currentView === 'module') {
        e.preventDefault();
        searchEnvInput.focus();
        searchEnvInput.select();
      }
    } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') {
      e.preventDefault();
      toggleSidebar();
    }
  });

  // Initial Boot
  initTheme();
  loadProjects();

  // Auto-launch tour on first visit
  if (!localStorage.getItem('simplyenv_tour_seen')) {
    setTimeout(() => openTourModal(0), 400);
  }
});
