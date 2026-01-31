import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { SHORTCUT_LIST } from '../hooks/useKeyboardShortcuts';

export function KeyboardShortcutsHelp() {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    const handleShow = () => setIsOpen(true);
    window.addEventListener('show-shortcuts-help', handleShow);
    return () => window.removeEventListener('show-shortcuts-help', handleShow);
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        setIsOpen(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen]);

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            className="shortcuts-backdrop"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setIsOpen(false)}
          />

          {/* Modal */}
          <motion.div
            className="shortcuts-modal"
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 20 }}
            transition={{ type: 'spring', damping: 25, stiffness: 300 }}
          >
            <div className="shortcuts-header">
              <h2>Keyboard Shortcuts</h2>
              <button className="shortcuts-close" onClick={() => setIsOpen(false)}>
                ✕
              </button>
            </div>

            <div className="shortcuts-content">
              {SHORTCUT_LIST.map((category) => (
                <div key={category.category} className="shortcuts-category">
                  <h3>{category.category}</h3>
                  <div className="shortcuts-list">
                    {category.shortcuts.map((shortcut, i) => (
                      <div key={i} className="shortcut-row">
                        <div className="shortcut-keys">
                          {shortcut.keys.map((key, j) => (
                            <span key={j}>
                              <kbd>{key}</kbd>
                              {j < shortcut.keys.length - 1 && <span className="key-plus">+</span>}
                            </span>
                          ))}
                        </div>
                        <span className="shortcut-desc">{shortcut.description}</span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="shortcuts-footer">
              <span className="shortcuts-hint">Press <kbd>?</kbd> to toggle this help</span>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
