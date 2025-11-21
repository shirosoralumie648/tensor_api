import type { Meta, StoryObj } from '@storybook/react';
import { Container } from './Container';

const meta = {
  title: 'Layout/Container',
  component: Container,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Container>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="bg-neutral-100">
      <Container>
        <div className="bg-white p-6 rounded-lg">
          <h2 className="text-2xl font-bold mb-2">Default Container</h2>
          <p className="text-neutral-600">This is the default container with xl max-width.</p>
        </div>
      </Container>
    </div>
  ),
};

export const Sizes: Story = {
  args: {
    children: (
      <div>
        <h2 className="text-2xl font-bold mb-4">Container Sizes</h2>
        <p className="text-neutral-600">Try resizing the browser to see different container sizes.</p>
      </div>
    ),
  },
};

export const Variants: Story = {
  render: () => (
    <div className="space-y-8">
      <div className="bg-blue-50">
        <Container size="sm">
          <div className="bg-white p-4 rounded">Small Container</div>
        </Container>
      </div>
      <div className="bg-green-50">
        <Container size="md">
          <div className="bg-white p-4 rounded">Medium Container</div>
        </Container>
      </div>
      <div className="bg-yellow-50">
        <Container size="lg">
          <div className="bg-white p-4 rounded">Large Container</div>
        </Container>
      </div>
      <div className="bg-red-50">
        <Container size="xl">
          <div className="bg-white p-4 rounded">XL Container</div>
        </Container>
      </div>
    </div>
  ),
};

export const NoPadding: Story = {
  args: {
    withPadding: false,
    children: (
      <div className="bg-primary-500 text-white h-32 flex items-center justify-center">
        No padding container (padding only on responsive breakpoints)
      </div>
    ),
  },
};

