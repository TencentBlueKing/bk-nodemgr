import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

describe('log-version dialog binding', () => {
  it('uses bkui Dialog is-show contract instead of plain v-model', () => {
    const source = readFileSync(resolve(__dirname, '../src/components/log-version.vue'), 'utf8');

    expect(source).toContain(':is-show="isShow"');
    expect(source).toContain('@closed="isShow = false"');
    expect(source).not.toContain('v-model="isShow"');
  });
});
