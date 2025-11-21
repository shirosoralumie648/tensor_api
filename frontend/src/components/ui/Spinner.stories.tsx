import type { Meta, StoryObj } from '@storybook/react';
import { Spinner } from './Spinner';

const meta = {
  title: 'UI/Spinner',
  component: Spinner,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Spinner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    size: 'md',
  },
};

export const Sizes: Story = {
  render: () => (
    <div className="flex gap-8">
      <Spinner size="sm" />
      <Spinner size="md" />
      <Spinner size="lg" />
      <Spinner size="xl" />
    </div>
  ),
};

export const Colors: Story = {
  render: () => (
    <div className="flex gap-8">
      <Spinner color="primary" />
      <Spinner color="white" className="bg-primary-500 p-4 rounded" />
    </div>
  ),
};

export const WithText: Story = {
  args: {
    size: 'lg',
    withText: 'Loading...',
  },
};

export const FullScreen: Story = {
  args: {
    fullScreen: true,
    withText: 'Please wait...',
  },
  parameters: {
    layout: 'fullscreen',
  },
};

export const Variants: Story = {
  render: () => (
    <div className="flex flex-col gap-8">
      <Spinner size="md" withText="Loading..." />
      <Spinner size="md" color="white" className="bg-primary-500 p-6 rounded" withText="Processing..." />
    </div>
  ),
};

