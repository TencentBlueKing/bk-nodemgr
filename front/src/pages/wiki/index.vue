<template>
  <section class="wiki-reader">
    <header class="reader-toolbar">
      <RouterLink :to="{ name: 'wiki', params: { language, document: ['README.md'] } }">
        {{ t('wiki.title') }}
      </RouterLink>
      <nav :aria-label="t('wiki.language')">
        <RouterLink
          v-for="item in languages"
          :key="item.code"
          :to="{ name: 'wiki', params: { language: item.code, document: route.params.document } }"
          :aria-current="language === item.code ? 'page' : undefined"
        >{{ item.label }}</RouterLink>
      </nav>
    </header>
    <Loading :loading="loading">
      <div v-if="error" class="reader-error" role="alert">
        <p>{{ error }}</p>
        <Button @click="loadDocument">{{ t('wiki.retry') }}</Button>
      </div>
      <div v-else-if="!loading" class="reader-body">
        <article id="wiki-reader-scroll" ref="article" class="reader-content" @click="navigate">
          <MdPreview
            id="wiki-reader"
            :model-value="content"
            preview-theme="github"
            :language="language === 'zh' ? 'zh-CN' : 'en-US'"
            :md-heading-id="headingId"
            :sanitize="prepareHTML"
            :code-style-reverse="false"
            no-mermaid
            no-katex
            no-echarts
            @on-remount="scrollToHeading"
          />
        </article>
        <aside class="reader-catalog" :aria-label="t('wiki.catalog')">
          <h2>{{ t('wiki.catalog') }}</h2>
          <MdCatalog editor-id="wiki-reader" scroll-element="#wiki-reader-scroll" :md-heading-id="headingId" />
        </aside>
      </div>
    </Loading>
  </section>
</template>

<script setup lang="ts">
import { Button, Loading } from 'bkui-vue';
import highlight from 'highlight.js/lib/common';
import { config, MdCatalog, MdPreview } from 'md-editor-v3';
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { contentURL, documentPath, embeddedPath, headingSlug } from './document';

import 'highlight.js/styles/github.css';
import 'md-editor-v3/lib/preview.css';

// Use local highlighting; enable neither remote extensions nor raw source HTML.
config({ editorExtensions: { highlight: { instance: highlight } } });

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const languages = [{ code: 'zh', label: '中文' }, { code: 'en', label: 'English' }];
const language = computed(() => String(route.params.language));
const path = computed(() => {
  const value = route.params.document;
  return Array.isArray(value) ? value.join('/') : value || '';
});
const content = ref('');
const loading = ref(false);
const error = ref('');
const article = ref<HTMLElement>();
const headingId = ({ index }: { index: number }) => `wiki-heading-${index}`;
let controller: AbortController | undefined;

function readingURL(url: URL): string | undefined {
  const path = embeddedPath(url);
  if (!path?.endsWith('.md')) return undefined;
  const [language, ...document] = path.split('/');
  return router.resolve({ name: 'wiki', params: { language, document }, hash: url.hash }).href;
}

// Source HTML is disabled by the renderer. Resolve links before inserting its output.
function prepareHTML(html: string): string {
  const template = document.createElement('template');
  template.innerHTML = html;
  const base = contentURL(`${language.value}/${path.value}`);
  template.content.querySelectorAll('a[href]').forEach((link) => {
    const href = link.getAttribute('href') || '';
    try {
      const url = new URL(href, base);
      if (!['http:', 'https:', 'mailto:'].includes(url.protocol)) {
        link.removeAttribute('href');
        return;
      }
      const reading = readingURL(url);
      link.setAttribute('href', reading || url.href);
      if (reading) link.setAttribute('data-wiki-link', '');
      else {
        link.setAttribute('target', '_blank');
        link.setAttribute('rel', 'noopener noreferrer');
      }
    } catch {
      link.removeAttribute('href');
    }
  });
  template.content.querySelectorAll('img').forEach((image) => {
    try {
      const url = new URL(image.getAttribute('src') || '', base);
      if (!embeddedPath(url)) {
        image.removeAttribute('src');
        return;
      }
      image.setAttribute('src', url.href);
    } catch {
      image.removeAttribute('src');
    }
  });
  return template.innerHTML;
}

function navigate(event: MouseEvent) {
  if (event.button !== 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
  const link = (event.target as Element).closest<HTMLAnchorElement>('a[data-wiki-link]');
  if (!link) return;
  event.preventDefault();
  const { hash } = new URL(link.href);
  void router.push(hash.slice(1));
  if (link.href === window.location.href) void scrollToHeading();
}

async function scrollToHeading() {
  await nextTick();
  if (!route.hash || !article.value) return;
  let fragment: string;
  try {
    fragment = decodeURIComponent(route.hash.slice(1));
  } catch {
    return;
  }
  const occurrences = new Map<string, number>();
  const headings = Array.from(article.value.querySelectorAll<HTMLElement>('h1, h2, h3, h4, h5, h6'));
  const target = headings.find((heading) => {
    const text = heading.textContent || '';
    const slug = headingSlug(text);
    const count = occurrences.get(slug) || 0;
    occurrences.set(slug, count + 1);
    return heading.id === fragment || text === fragment || `${slug}${count ? `-${count}` : ''}` === fragment;
  });
  if (target) article.value.scrollTo({ top: target.offsetTop, behavior: 'instant' });
}

async function loadDocument() {
  controller?.abort();
  const request = new AbortController();
  controller = request;
  loading.value = true;
  error.value = '';
  content.value = '';
  const name = documentPath(language.value, path.value);
  if (!name) {
    error.value = t('wiki.notFound');
    loading.value = false;
    return;
  }
  try {
    const response = await fetch(contentURL(name), { credentials: 'same-origin', signal: request.signal });
    if (!response.ok) {
      if (response.status === 404) throw new Error(t('wiki.notFound'));
      throw new Error(`HTTP ${response.status}`);
    }
    if (!/^text\/(plain|markdown)\b/i.test(response.headers.get('content-type') || '')) {
      throw new Error(t('wiki.invalidResponse'));
    }
    const text = await response.text();
    if (!request.signal.aborted) content.value = text;
  } catch (cause) {
    if (request.signal.aborted) return;
    error.value = `${t('wiki.loadFailed')}: ${cause instanceof Error ? cause.message : String(cause)}`;
  } finally {
    if (controller === request) loading.value = false;
  }
}

watch([language, path], () => void loadDocument(), { immediate: true });
watch(() => route.hash, () => void scrollToHeading());
onBeforeUnmount(() => controller?.abort());
</script>

<style scoped>
.wiki-reader {
  padding: 24px;
}

.reader-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 1280px;
  margin: 0 auto 16px;
  font-size: 16px;
}

.reader-toolbar nav {
  display: flex;
  gap: 16px;
  font-size: 14px;
}

.reader-toolbar a {
  color: #63656e;
}

.reader-toolbar a[aria-current='page'] {
  color: #3a84ff;
}

.reader-body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 240px;
  max-width: 1280px;
  margin: 0 auto;
  background: #fff;
  border: 1px solid #dcdee5;
}

.reader-content {
  position: relative;
  min-width: 0;
  height: calc(100vh - 170px);
  min-height: 360px;
  overflow: auto;
  padding: 16px 32px;
}

.reader-catalog {
  max-height: calc(100vh - 170px);
  overflow: auto;
  padding: 24px 16px;
  border-left: 1px solid #eaebf0;
}

.reader-catalog h2 {
  margin-bottom: 16px;
  font-size: 14px;
}

/* Restore document markers removed by the application's global CSS reset. */
.reader-content :deep(.md-editor-preview ol) {
  list-style-type: decimal;
}

.reader-content :deep(.md-editor-preview ul) {
  list-style-type: disc;
}

.reader-content :deep(.md-editor-preview .contains-task-list) {
  list-style-type: none;
}

.reader-catalog :deep(.md-editor-catalog-active > span),
.reader-catalog :deep(.md-editor-catalog-link span:hover) {
  color: #3a84ff;
}

.reader-catalog :deep(.md-editor-catalog-indicator) {
  background-color: #3a84ff;
}

.reader-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 48px;
  background: #fff;
  color: #63656e;
}

@media (max-width: 900px) {
  .reader-body {
    grid-template-columns: minmax(0, 1fr);
  }

  .reader-catalog {
    display: none;
  }

  .reader-content {
    padding: 16px;
  }
}
</style>
