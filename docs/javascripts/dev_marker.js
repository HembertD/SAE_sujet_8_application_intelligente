document.addEventListener('DOMContentLoaded', function () {
  const siteName = document.querySelector('.md-header__title, .md-header__topic, .md-tabs__link');
  if (!siteName) return;

  const text = siteName.textContent || '';
  if (!text.includes('*')) return;

  const marker = document.createElement('span');
  marker.className = 'dev-marker';
  marker.setAttribute('tabindex', '0');
  marker.setAttribute('data-tooltip', 'Documentation en cours de développement');
  marker.textContent = '*';

  const current = siteName.textContent.trim();
  siteName.innerHTML = '';
  siteName.appendChild(document.createTextNode(current.replace(/\*\s*$/, '')));
  siteName.appendChild(marker);
});
