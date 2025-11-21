import type { Meta, StoryObj } from '@storybook/react';
import { Card, CardHeader, CardBody, CardFooter } from './Card';
import { Button } from './Button';

const meta = {
  title: 'UI/Card',
  component: Card,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    hoverable: {
      control: 'boolean',
    },
    bordered: {
      control: 'boolean',
    },
    shadow: {
      control: 'select',
      options: ['none', 'sm', 'md', 'lg', 'xl'],
    },
  },
} satisfies Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: <p>This is a basic card with some content.</p>,
  },
};

export const WithTitle: Story = {
  args: {
    children: (
      <div>
        <h3 className="text-lg font-semibold mb-2">Card Title</h3>
        <p className="text-neutral-600">This is the card content.</p>
      </div>
    ),
  },
};

export const Hoverable: Story = {
  args: {
    hoverable: true,
    children: (
      <div>
        <h3 className="text-lg font-semibold mb-2">Hoverable Card</h3>
        <p className="text-neutral-600">Hover over this card to see the effect.</p>
      </div>
    ),
  },
};

export const WithoutBorder: Story = {
  args: {
    bordered: false,
    shadow: 'md',
    children: (
      <div>
        <h3 className="text-lg font-semibold mb-2">Card without Border</h3>
        <p className="text-neutral-600">This card has no border, only shadow.</p>
      </div>
    ),
  },
};

export const Shadows: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Card shadow="none">
        <p className="font-medium">No Shadow</p>
      </Card>
      <Card shadow="sm">
        <p className="font-medium">Small Shadow</p>
      </Card>
      <Card shadow="md">
        <p className="font-medium">Medium Shadow</p>
      </Card>
      <Card shadow="lg">
        <p className="font-medium">Large Shadow</p>
      </Card>
      <Card shadow="xl">
        <p className="font-medium">Extra Large Shadow</p>
      </Card>
    </div>
  ),
};

export const WithHeaderBodyFooter: Story = {
  render: () => (
    <Card className="w-full max-w-md">
      <CardHeader>
        <h2 className="text-xl font-bold">Card Title</h2>
      </CardHeader>
      <CardBody>
        <p className="text-neutral-600">
          This is the card body content. It can contain any type of content you want to display.
        </p>
      </CardBody>
      <CardFooter>
        <div className="flex gap-2 justify-end">
          <Button variant="ghost">Cancel</Button>
          <Button>Confirm</Button>
        </div>
      </CardFooter>
    </Card>
  ),
};

export const Grid: Story = {
  render: () => (
    <div className="grid grid-cols-3 gap-4">
      {[1, 2, 3, 4, 5, 6].map((i) => (
        <Card key={i} hoverable>
          <h4 className="font-semibold mb-2">Card {i}</h4>
          <p className="text-sm text-neutral-600">This is card number {i}</p>
        </Card>
      ))}
    </div>
  ),
};

export const ComplexContent: Story = {
  render: () => (
    <Card className="w-full max-w-2xl">
      <CardHeader>
        <h2 className="text-2xl font-bold">User Profile</h2>
      </CardHeader>
      <CardBody>
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-primary-100 flex items-center justify-center">
              <span className="text-2xl">👤</span>
            </div>
            <div>
              <h3 className="font-semibold">John Doe</h3>
              <p className="text-sm text-neutral-500">john@example.com</p>
            </div>
          </div>
          <div className="border-t pt-4">
            <p className="text-sm text-neutral-600">
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
              incididunt ut labore et dolore magna aliqua.
            </p>
          </div>
        </div>
      </CardBody>
      <CardFooter>
        <div className="flex gap-2 justify-end">
          <Button variant="outline">Edit</Button>
          <Button>Save</Button>
        </div>
      </CardFooter>
    </Card>
  ),
};

