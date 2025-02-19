import { describe, expect, it } from 'vitest';

import { mount } from '@vue/test-utils';

import FlexRow from '../src/components/flex-row.vue';

describe('FlexRow.vue', () => {
  it('renders correctly', () => {
    const wrapper = mount(FlexRow);
    expect(wrapper.exists()).toBe(true);
  });

  it('has two div elements', () => {
    const wrapper = mount(FlexRow);
    expect(wrapper.findAll('div')).toHaveLength(3); // One parent div and two child divs
  });

  it('renders left slot content', () => {
    const wrapper = mount(FlexRow, {
      slots: {
        left: '<span>Left Content</span>',
      },
    });
    expect(wrapper.find('div > div:first-child').html()).toContain('Left Content');
  });

  it('renders right slot content', () => {
    const wrapper = mount(FlexRow, {
      slots: {
        right: '<span>Right Content</span>',
      },
    });
    expect(wrapper.find('div > div:last-child').html()).toContain('Right Content');
  });

  it('has correct CSS classes', () => {
    const wrapper = mount(FlexRow);
    expect(wrapper.classes()).toContain('flex');
    expect(wrapper.classes()).toContain('items-center');
    expect(wrapper.classes()).toContain('place-content-between');
  });
});
