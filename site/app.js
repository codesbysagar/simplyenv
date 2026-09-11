// simplyenv Product Website Controller
document.addEventListener('DOMContentLoaded', () => {
  // Theme Management
  const themeToggleBtn = document.getElementById('themeToggleBtn');
  const themeMoonIcon = document.getElementById('themeMoonIcon');
  const themeSunIcon = document.getElementById('themeSunIcon');

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
      localStorage.setItem('simplyenv_site_theme', next);
      updateThemeIcons(next);
    };
  }

  // Install Command Tabs & Binaries Switcher
  const installCommands = {
    curl: 'curl -fsSL https://simplyenv.codesbysagar.com/install.sh | bash',
    go: 'go install github.com/codesbysagar/simplyenv/cmd/simplyenv@latest',
    git: 'git clone https://github.com/codesbysagar/simplyenv.git && cd simplyenv && go build -o simplyenv ./cmd/simplyenv'
  };

  const installCodeBox = document.getElementById('installCodeBox');
  const installBinariesBox = document.getElementById('installBinariesBox');
  const installCommandText = document.getElementById('installCommandText');
  const installTabs = document.querySelectorAll('.install-tab');

  installTabs.forEach(tab => {
    tab.onclick = () => {
      installTabs.forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      const tabKey = tab.getAttribute('data-tab');

      if (tabKey === 'binaries') {
        if (installCodeBox) installCodeBox.classList.add('hidden');
        if (installBinariesBox) installBinariesBox.classList.remove('hidden');
      } else {
        if (installBinariesBox) installBinariesBox.classList.add('hidden');
        if (installCodeBox) installCodeBox.classList.remove('hidden');
        if (installCommands[tabKey]) {
          installCommandText.textContent = installCommands[tabKey];
        }
      }
    };
  });

  // Copy Install Command
  const copyInstallBtn = document.getElementById('copyInstallBtn');
  if (copyInstallBtn) {
    copyInstallBtn.onclick = () => {
      const text = installCommandText.textContent;
      navigator.clipboard.writeText(text).then(() => {
        const label = copyInstallBtn.querySelector('.copy-label');
        const original = label.textContent;
        label.textContent = 'Copied!';
        copyInstallBtn.style.borderColor = 'var(--accent-emerald)';
        setTimeout(() => {
          label.textContent = original;
          copyInstallBtn.style.borderColor = '';
        }, 2000);
      });
    };
  }

  // Showcase Frame Light / Dark Mockup Switcher
  const showcaseLightBtn = document.getElementById('showcaseLightBtn');
  const showcaseDarkBtn = document.getElementById('showcaseDarkBtn');
  const showcaseImg = document.getElementById('showcaseImg');

  if (showcaseLightBtn && showcaseDarkBtn && showcaseImg) {
    showcaseLightBtn.onclick = () => {
      showcaseLightBtn.classList.add('active');
      showcaseDarkBtn.classList.remove('active');
      showcaseImg.src = '/assets/dashboard_light_mode_1789106014791.png';
    };

    showcaseDarkBtn.onclick = () => {
      showcaseDarkBtn.classList.add('active');
      showcaseLightBtn.classList.remove('active');
      showcaseImg.src = '/assets/dashboard_dark_mode_1789106037988.png';
    };
  }

  // Docs Method Selector (Step 1 Installation Options)
  const docsMethodBtns = document.querySelectorAll('.docs-method-btn');
  const docsPanels = {
    curl: document.getElementById('docsPanelCurl'),
    binaries: document.getElementById('docsPanelBinaries'),
    go: document.getElementById('docsPanelGo')
  };

  docsMethodBtns.forEach(btn => {
    btn.onclick = () => {
      docsMethodBtns.forEach(b => {
        b.classList.remove('active');
        b.setAttribute('aria-selected', 'false');
      });
      btn.classList.add('active');
      btn.setAttribute('aria-selected', 'true');

      const method = btn.getAttribute('data-method');
      Object.entries(docsPanels).forEach(([key, panel]) => {
        if (panel) {
          if (key === method) {
            panel.classList.remove('hidden');
          } else {
            panel.classList.add('hidden');
          }
        }
      });
    };
  });

  // Copy Panel Buttons in Docs
  document.querySelectorAll('.copy-panel-btn').forEach(btn => {
    btn.onclick = () => {
      const textToCopy = btn.getAttribute('data-copy');
      if (textToCopy) {
        navigator.clipboard.writeText(textToCopy).then(() => {
          const span = btn.querySelector('span');
          const original = span ? span.textContent : 'Copy';
          if (span) span.textContent = 'Copied!';
          btn.style.borderColor = 'var(--accent-emerald)';
          btn.style.color = 'var(--accent-emerald)';
          setTimeout(() => {
            if (span) span.textContent = original;
            btn.style.borderColor = '';
            btn.style.color = '';
          }, 2000);
        });
      }
    };
  });

  // Initialize
  initTheme();
});
