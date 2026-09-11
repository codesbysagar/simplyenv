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

  // Install Command Tabs
  const installCommands = {
    go: 'go install github.com/codesbysagar/simplyenv/cmd/simplyenv@latest',
    git: 'git clone https://github.com/codesbysagar/simplyenv.git && cd simplyenv && go build -o simplyenv cmd/simplyenv/main.go'
  };

  const installCommandText = document.getElementById('installCommandText');
  const installTabs = document.querySelectorAll('.install-tab');

  installTabs.forEach(tab => {
    tab.onclick = () => {
      installTabs.forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      const tabKey = tab.getAttribute('data-tab');
      if (installCommands[tabKey]) {
        installCommandText.textContent = installCommands[tabKey];
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

  // Initialize
  initTheme();
});
