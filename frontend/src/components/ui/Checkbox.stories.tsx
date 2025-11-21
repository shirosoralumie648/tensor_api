import type { Meta, StoryObj } from '@storybook/react';
import { Checkbox } from './Checkbox';
import { useState } from 'react';

const meta = {
  title: 'UI/Checkbox',
  component: Checkbox,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    label: 'Accept terms and conditions',
  },
};

export const Checked: Story = {
  args: {
    label: 'Remember me',
    defaultChecked: true,
  },
};

export const Disabled: Story = {
  args: {
    label: 'Disabled checkbox',
    disabled: true,
  },
};

export const WithHelperText: Story = {
  args: {
    label: 'Subscribe to updates',
    helperText: 'We will send you important updates',
  },
};

export const WithError: Story = {
  args: {
    label: 'Agree to terms',
    error: 'You must agree to continue',
  },
};

export const MultipleCheckboxes: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox label="Option 1" />
      <Checkbox label="Option 2" />
      <Checkbox label="Option 3" />
      <Checkbox label="Option 4" defaultChecked />
    </div>
  ),
};

