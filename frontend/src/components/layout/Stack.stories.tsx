import type { Meta, StoryObj } from '@storybook/react';
import { Stack, VStack, HStack } from './Stack';

const meta = {
  title: 'Layout/Stack',
  component: Stack,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Stack>;

export default meta;
type Story = StoryObj<typeof meta>;

const Item = ({ children }: any) => (
  <div className="bg-primary-500 text-white p-4 rounded">
    {children}
  </div>
);

export const ColumnDefault: Story = {
  render: () => (
    <VStack spacing="md" className="w-96">
      <Item>Item 1</Item>
      <Item>Item 2</Item>
      <Item>Item 3</Item>
    </VStack>
  ),
};

export const RowDefault: Story = {
  render: () => (
    <HStack spacing="md">
      <Item>Item 1</Item>
      <Item>Item 2</Item>
      <Item>Item 3</Item>
    </HStack>
  ),
};

export const Spacing: Story = {
  render: () => (
    <div className="space-y-8">
      <VStack spacing="xs" className="w-96">
        <Item>XS</Item>
        <Item>XS</Item>
      </VStack>
      <VStack spacing="sm" className="w-96">
        <Item>SM</Item>
        <Item>SM</Item>
      </VStack>
      <VStack spacing="md" className="w-96">
        <Item>MD</Item>
        <Item>MD</Item>
      </VStack>
      <VStack spacing="lg" className="w-96">
        <Item>LG</Item>
        <Item>LG</Item>
      </VStack>
      <VStack spacing="xl" className="w-96">
        <Item>XL</Item>
        <Item>XL</Item>
      </VStack>
    </div>
  ),
};

export const Alignment: Story = {
  render: () => (
    <div className="space-y-8">
      <HStack spacing="md" align="start" fullWidth>
        <Item>Start</Item>
        <Item>Higher</Item>
      </HStack>
      <HStack spacing="md" align="center" fullWidth>
        <Item>Center</Item>
        <Item>Center</Item>
      </HStack>
      <HStack spacing="md" align="end" fullWidth>
        <Item>End</Item>
        <Item>End</Item>
      </HStack>
    </div>
  ),
};

export const Justification: Story = {
  render: () => (
    <div className="space-y-8">
      <HStack spacing="md" justify="start" fullWidth className="bg-neutral-100 p-4">
        <Item>Start</Item>
        <Item>Start</Item>
      </HStack>
      <HStack spacing="md" justify="center" fullWidth className="bg-neutral-100 p-4">
        <Item>Center</Item>
        <Item>Center</Item>
      </HStack>
      <HStack spacing="md" justify="end" fullWidth className="bg-neutral-100 p-4">
        <Item>End</Item>
        <Item>End</Item>
      </HStack>
      <HStack spacing="md" justify="between" fullWidth className="bg-neutral-100 p-4">
        <Item>Between</Item>
        <Item>Between</Item>
      </HStack>
    </div>
  ),
};

