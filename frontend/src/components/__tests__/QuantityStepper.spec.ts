import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import QuantityStepper from '../ui/QuantityStepper.vue'
import { i18n } from '@/i18n'

function factory(props: { modelValue: number; min?: number; max?: number }) {
  const wrapper = mount(QuantityStepper, {
    props: {
      ...props,
      'onUpdate:modelValue': (v: number) => {
        void wrapper.setProps({ modelValue: v })
      },
    },
    global: { plugins: [i18n] },
  })
  return wrapper
}

describe('QuantityStepper', () => {
  it('increments and decrements within bounds', async () => {
    const wrapper = factory({ modelValue: 1, min: 1, max: 3 })
    const [dec, inc] = wrapper.findAll('button')
    expect(dec!.attributes('disabled')).toBeDefined()
    await inc!.trigger('click')
    expect(wrapper.props('modelValue')).toBe(2)
    await inc!.trigger('click')
    expect(wrapper.props('modelValue')).toBe(3)
    expect(inc!.attributes('disabled')).toBeDefined()
    await dec!.trigger('click')
    expect(wrapper.props('modelValue')).toBe(2)
  })

  it('clamps typed values to max', async () => {
    const wrapper = factory({ modelValue: 1, min: 1, max: 5 })
    await wrapper.find('input').setValue('50')
    expect(wrapper.props('modelValue')).toBe(5)
  })

  it('has accessible labels on buttons', () => {
    const wrapper = factory({ modelValue: 1 })
    const labels = wrapper.findAll('button').map((b) => b.attributes('aria-label'))
    expect(labels).toEqual(['Decrease quantity', 'Increase quantity'])
  })
})
