import type { Meta, StoryObj } from '@storybook/react';
import { Alert } from './Alert';

const meta = {
  title: 'UI/Alert',
  component: Alert,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Success: Story = {
  args: {
    variant: 'success',
    title: 'Success!',
    description: 'Your operation has been completed successfully.',
  },
};

export const Error: Story = {
  args: {
    variant: 'error',
    title: 'Error',
    description: 'Something went wrong. Please try again.',
  },
};

export const Warning: Story = {
  args: {
    variant: 'warning',
    title: 'Warning',
    description: 'Please be careful with this action.',
  },
};

export const Info: Story = {
  args: {
    variant: 'info',
    title: 'Information',
    description: 'This is an informational message.',
  },
};

export const Closable: Story = {
  args: {
    variant: 'info',
    title: 'Closable Alert',
    description: 'Click the X button to close this alert.',
    closable: true,
  },
};

export const WithChildren: Story = {
  args: {
    variant: 'warning',
    title: 'Complex Alert',
    closable: true,
    children: (
      <div>
        <p>This is a more complex alert with multiple lines of text.</p>
        <p>You can include any content you want here.</p>
      </div>
    ),
  },
};

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-96">
      <Alert variant="success" title="Success" description="Operation completed" />
      <Alert variant="error" title="Error" description="Something went wrong" />
      <Alert variant="warning" title="Warning" description="Please be careful" />
      <Alert variant="info" title="Info" description="Here is some information" />
    </div>
  ),
};

