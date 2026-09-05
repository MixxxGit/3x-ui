import { cn } from '@/lib/cn';

// Set when the site is served from a subpath (GitHub Pages project site).
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || '';

// Official 3x-ui logo (media/3x-ui-{light,dark}.png from the upstream repo).
// Theme-aware via Tailwind's `dark:` variant. Pass a height class (e.g. `h-6`);
// width scales automatically (the artwork is 2:1).
export function Logo({ className }: { className?: string }) {
  return (
    <>
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={`${basePath}/logo-light.png`}
        alt="3x-ui"
        className={cn('w-auto dark:hidden', className)}
      />
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={`${basePath}/logo-dark.png`}
        alt="3x-ui"
        className={cn('hidden w-auto dark:block', className)}
      />
    </>
  );
}
