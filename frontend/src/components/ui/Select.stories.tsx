import type { Meta, StoryObj } from '@storybook/react';
import { Select } from './Select';

const meta = {
  title: 'UI/Select',
  component: Select,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

const options = [
  { value: '1', label: 'Option 1' },
  { value: '2', label: 'Option 2' },
  { value: '3', label: 'Option 3' },
  { value: '4', label: 'Option 4' },
];

export const Default: Story = {
  args: {
    options,
    placeholder: 'Select an option',
  },
};

export const WithLabel: Story = {
  args: {
    label: 'Choose Category',
    options,
    placeholder: 'Select a category',
  },
};

export const WithHelperText: Story = {
  args: {
    label: 'Country',
    options,
    helperText: 'Please select your country',
    placeholder: 'Select country',
  },
};

export const WithError: Story = {
  args: {
    label: 'Required Field',
    error: 'This field is required',
    options,
    required: true,
  },
};

export const Disabled: Story = {
  args: {
    label: 'Disabled Select',
    options,
    disabled: true,
    placeholder: 'This is disabled',
  },
};

