import type { Meta, StoryObj } from '@storybook/react';
import { Progress } from './Progress';

const meta = {
  title: 'UI/Progress',
  component: Progress,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Progress>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    value: 50,
    max: 100,
  },
};

export const WithLabel: Story = {
  args: {
    value: 65,
    max: 100,
    label: 'Upload Progress',
  },
};

export const WithPercentage: Story = {
  args: {
    value: 75,
    max: 100,
    showPercentage: true,
  },
};

export const WithLabelAndPercentage: Story = {
  args: {
    value: 85,
    max: 100,
    label: 'Download Progress',
    showPercentage: true,
  },
};

export const Sizes: Story = {
  args: { value: 50 },
  render: () => (
    <div className="flex flex-col gap-6 w-96">
      <Progress value={50} size="sm" label="Small" showPercentage />
      <Progress value={50} size="md" label="Medium" showPercentage />
      <Progress value={50} size="lg" label="Large" showPercentage />
    </div>
  ),
};

export const Colors: Story = {
  args: { value: 60 },
  render: () => (
    <div className="flex flex-col gap-6 w-96">
      <Progress value={60} color="primary" label="Primary" showPercentage />
      <Progress value={60} color="success" label="Success" showPercentage />
      <Progress value={60} color="warning" label="Warning" showPercentage />
      <Progress value={60} color="error" label="Error" showPercentage />
    </div>
  ),
};

export const Complete: Story = {
  args: {
    value: 100,
    max: 100,
    label: 'Complete',
    showPercentage: true,
    color: 'success',
  },
};

