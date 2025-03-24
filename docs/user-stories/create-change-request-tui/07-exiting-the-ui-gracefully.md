## 7. Exiting the UI Gracefully

**User Story**  
As a CLI user,  
I want to exit the interface at any time using a known key,  
so that I can safely cancel the operation.

**Acceptance Criteria**
- Pressing `ESC`:
  - Cancels the operation
  - Displays: `Change request creation canceled by user.`
  - Discards all current selections