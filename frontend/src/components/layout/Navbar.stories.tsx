import type { Meta, StoryObj } from '@storybook/react';
import {
  Navbar,
  NavbarContent,
  NavbarBrand,
  NavbarMenu,
  NavbarItem,
  NavbarActions,
} from './Navbar';
import { Button } from '../ui/Button';
import { Menu } from 'lucide-react';

const meta = {
  title: 'Layout/Navbar',
  component: Navbar,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Navbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Navbar>
      <NavbarContent>
        <NavbarBrand>
          <div className="text-2xl font-bold text-primary-500">Logo</div>
        </NavbarBrand>
        <NavbarMenu>
          <NavbarItem href="#" active>
            Home
          </NavbarItem>
          <NavbarItem href="#">About</NavbarItem>
          <NavbarItem href="#">Services</NavbarItem>
          <NavbarItem href="#">Contact</NavbarItem>
        </NavbarMenu>
        <NavbarActions>
          <Button variant="outline" size="sm">
            Login
          </Button>
          <Button size="sm">
            Sign Up
          </Button>
        </NavbarActions>
      </NavbarContent>
    </Navbar>
  ),
};

export const WithoutBorder: Story = {
  render: () => (
    <Navbar bordered={false}>
      <NavbarContent>
        <NavbarBrand>
          <div className="text-2xl font-bold text-primary-500">Oblivious</div>
        </NavbarBrand>
        <NavbarMenu>
          <NavbarItem href="#" active>
            Dashboard
          </NavbarItem>
          <NavbarItem href="#">API</NavbarItem>
          <NavbarItem href="#">Docs</NavbarItem>
        </NavbarMenu>
        <NavbarActions>
          <Button size="sm">Start Free</Button>
        </NavbarActions>
      </NavbarContent>
    </Navbar>
  ),
};

export const Sticky: Story = {
  render: () => (
    <div className="h-96">
      <Navbar sticky>
        <NavbarContent>
          <NavbarBrand>
            <div className="text-xl font-bold">Sticky Nav</div>
          </NavbarBrand>
          <NavbarMenu>
            <NavbarItem href="#">Item 1</NavbarItem>
            <NavbarItem href="#">Item 2</NavbarItem>
          </NavbarMenu>
        </NavbarContent>
      </Navbar>
      <div className="p-8 bg-neutral-50">
        <p>Scroll down to see the sticky navbar</p>
        <div className="space-y-4 mt-4">
          {Array.from({ length: 10 }).map((_, i) => (
            <div key={i} className="bg-white p-4 rounded">
              Content {i + 1}
            </div>
          ))}
        </div>
      </div>
    </div>
  ),
};

export const Complex: Story = {
  render: () => (
    <Navbar>
      <NavbarContent>
        <NavbarBrand>
          <div className="text-2xl font-bold bg-gradient-to-r from-primary-500 to-blue-500 bg-clip-text text-transparent">
            Oblivious
          </div>
        </NavbarBrand>
        <NavbarMenu>
          <NavbarItem href="#" active>
            Explore
          </NavbarItem>
          <NavbarItem href="#">Templates</NavbarItem>
          <NavbarItem href="#">Pricing</NavbarItem>
          <NavbarItem href="#">Blog</NavbarItem>
        </NavbarMenu>
        <NavbarActions>
          <Button variant="ghost" size="sm">
            GitHub
          </Button>
          <Button variant="outline" size="sm">
            Log in
          </Button>
          <Button size="sm">
            Get Started
          </Button>
        </NavbarActions>
      </NavbarContent>
    </Navbar>
  ),
};

