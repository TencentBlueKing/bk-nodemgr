// Keep document and resource paths inside the application's embedded Wiki.
export function documentPath(language: string, path: string): string | undefined {
  if (!['zh', 'en'].includes(language) || !path.endsWith('.md')) return undefined;
  const segments = path.split('/');
  if (segments.some(segment => !segment || segment === '.' || segment === '..' || /[\\\u0000-\u001f]/.test(segment))) {
    return undefined;
  }
  return `${language}/${path}`;
}

export function contentURL(path: string): URL {
  const root = new URL('wiki/', window.location.href);
  return new URL(path.split('/').map(encodeURIComponent)
    .join('/'), root);
}

export function embeddedPath(url: URL): string | undefined {
  const root = new URL('wiki/', window.location.href);
  if (url.origin !== root.origin || !url.pathname.startsWith(root.pathname)) return undefined;
  try {
    const path = decodeURIComponent(url.pathname.slice(root.pathname.length));
    const segments = path.split('/');
    if (!['zh', 'en'].includes(segments[0]) || segments.length < 2
      || segments.some(segment => !segment || segment === '.' || segment === '..' || /[\\\u0000-\u001f]/.test(segment))) {
      return undefined;
    }
    return path;
  } catch {
    return undefined;
  }
}

// Match common Markdown heading fragments while keeping duplicate headings distinct.
export function headingSlug(text: string): string {
  return text.trim().toLowerCase()
    .replace(/[^\p{L}\p{N}\p{M}\s_-]/gu, '')
    .replace(/\s/g, '-');
}
