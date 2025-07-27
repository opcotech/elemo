import type { Transition, Variants } from "framer-motion";
import { useReducedMotion } from "framer-motion";

// Button Animations
export const buttonVariants: Variants = {
  initial: {
    transition: { type: "spring", stiffness: 500, damping: 30 },
  },
  hover: {
    transition: { type: "spring", stiffness: 400, damping: 22 },
  },
  tap: {
    scale: 0.98,
    transition: { type: "spring", stiffness: 3000, damping: 12, mass: 0.3 },
  },
  focus: {
    transition: { type: "spring", stiffness: 400, damping: 24 },
  },
};

// Modal (Dialog) Animations
export const modalVariants: Variants = {
  hidden: {
    opacity: 0,
    scale: 0.8,
    y: -24,
    transition: { duration: 0.05, ease: "easeOut" },
  },
  visible: {
    opacity: 1,
    scale: 1,
    y: 0,
    transition: { duration: 0.05, ease: "easeOut" },
  },
  exit: {
    opacity: 0,
    scale: 0.8,
    y: -24,
    transition: { duration: 0.05, ease: "easeIn" },
  },
};

// Modal Backdrop Animations
export const backdropVariants: Variants = {
  hidden: { opacity: 0, transition: { duration: 0.1 } },
  visible: { opacity: 1, transition: { duration: 0.15 } },
  exit: { opacity: 0, transition: { duration: 0.1 } },
};

// Card Animations
export const cardVariants: Variants = {
  initial: {
    y: 16,
    opacity: 0,
    scale: 0.98,
    transition: { type: "spring", stiffness: 200, damping: 30 },
  },
  animate: {
    y: 0,
    opacity: 1,
    scale: 1,
    transition: { type: "spring", stiffness: 200, damping: 24 },
  },
  hover: {
    y: -4,
    scale: 1.015,
    boxShadow: "0 4px 16px 0 rgba(0,0,0,0.08)",
    transition: { type: "spring", stiffness: 300, damping: 22 },
  },
};

// Utility hook for reduced motion-aware transitions
export function useMotionSafeTransition(transition: Transition): Transition {
  const reducedMotion = useReducedMotion();
  if (reducedMotion) {
    // Remove duration, spring, and other motion for accessibility
    return { ...transition, duration: 0 };
  }
  return transition;
}
