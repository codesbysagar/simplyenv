// simplyenv Product Website Controller — Retroverse Edition
document.addEventListener('DOMContentLoaded', () => {
  // Theme Management (Light / Dark Mode)
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
        const original = label ? label.textContent : 'Copy';
        if (label) label.textContent = 'Copied!';
        copyInstallBtn.style.backgroundColor = 'var(--color-lime)';
        copyInstallBtn.style.color = '#030100';
        copyInstallBtn.style.borderColor = '#030100';
        setTimeout(() => {
          if (label) label.textContent = original;
          copyInstallBtn.style.backgroundColor = '';
          copyInstallBtn.style.color = '';
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
          btn.style.backgroundColor = 'var(--color-lime)';
          btn.style.color = '#030100';
          btn.style.borderColor = '#030100';
          setTimeout(() => {
            if (span) span.textContent = original;
            btn.style.backgroundColor = '';
            btn.style.color = '';
            btn.style.borderColor = '';
          }, 2000);
        });
      }
    };
  });

  // Mobile Navigation Drawer
  const mobileMenuBtn = document.getElementById('mobileMenuBtn');
  const mobileNavDrawer = document.getElementById('mobileNavDrawer');
  const mobileNavBackdrop = document.getElementById('mobileNavBackdrop');

  if (mobileMenuBtn && mobileNavDrawer) {
    const hamburgerIcon = mobileMenuBtn.querySelector('.hamburger-icon');
    const closeIcon = mobileMenuBtn.querySelector('.close-icon');

    function openMobileMenu() {
      mobileNavDrawer.classList.add('open');
      if (mobileNavBackdrop) mobileNavBackdrop.classList.add('open');
      mobileMenuBtn.setAttribute('aria-expanded', 'true');
      if (hamburgerIcon) hamburgerIcon.classList.add('hidden');
      if (closeIcon) closeIcon.classList.remove('hidden');
      document.body.style.overflow = 'hidden';
    }

    function closeMobileMenu() {
      mobileNavDrawer.classList.remove('open');
      if (mobileNavBackdrop) mobileNavBackdrop.classList.remove('open');
      mobileMenuBtn.setAttribute('aria-expanded', 'false');
      if (hamburgerIcon) hamburgerIcon.classList.remove('hidden');
      if (closeIcon) closeIcon.classList.add('hidden');
      document.body.style.overflow = '';
    }

    mobileMenuBtn.onclick = () => {
      const isOpen = mobileNavDrawer.classList.contains('open');
      if (isOpen) {
        closeMobileMenu();
      } else {
        openMobileMenu();
      }
    };

    if (mobileNavBackdrop) {
      mobileNavBackdrop.onclick = closeMobileMenu;
    }

    const mobileNavCloseBtn = document.getElementById('mobileNavCloseBtn');
    if (mobileNavCloseBtn) {
      mobileNavCloseBtn.onclick = closeMobileMenu;
    }

    // Close on link click inside drawer
    mobileNavDrawer.querySelectorAll('a').forEach(link => {
      link.addEventListener('click', () => {
        closeMobileMenu();
      });
    });

    // Close on Escape
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && mobileNavDrawer.classList.contains('open')) {
        closeMobileMenu();
      }
    });

    // Close on resize > 768px
    window.addEventListener('resize', () => {
      if (window.innerWidth > 768 && mobileNavDrawer.classList.contains('open')) {
        closeMobileMenu();
      }
    });
  }

  // Initialize Theme
  initTheme();
});
