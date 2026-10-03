import { useEffect, useState } from 'react';

// stripHtml removes all HTML tags from a string and returns plain text.
// RSS feed descriptions frequently contain HTML. We do not trust that content
// and rendering it via dangerouslySetInnerHTML creates an XSS vector.
// DOMParser is available in all modern browsers and performs the same parsing
// the browser would do, but we only extract textContent — never innerHTML.
function stripHtml(html: string): string {
  const doc = new DOMParser().parseFromString(html, 'text/html');
  return doc.body.textContent ?? '';
}

// Mirrors ContentItem from the Go backend
interface ContentItem {
  ID: number;
  SourceID: number;
  Title: string;
  URL: string;
  ContentType: string;
  Summary?: string;
  WhyItMatters?: string;
  PersonalizationExplanation?: string;
  Author?: string;
  PublishedAt?: string;
}

// Mirrors Source from the Go backend
interface Source {
  ID: number;
  Name: string;
}

function App() {
  const [items, setItems] = useState<ContentItem[]>([]);
  const [sources, setSources] = useState<Record<number, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Track items that the user has interacted with to optimistically hide them
  const [hiddenItemIds, setHiddenItemIds] = useState<Set<number>>(new Set());

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        // Fetch both sources (for names) and feed items
        const [feedRes, sourcesRes] = await Promise.all([
          fetch('/api/feed'),
          fetch('/api/sources')
        ]);

        if (!feedRes.ok) throw new Error('Failed to fetch feed');
        if (!sourcesRes.ok) throw new Error('Failed to fetch sources');

        const feedData: ContentItem[] = await feedRes.json() || [];
        const sourcesData: Source[] = await sourcesRes.json() || [];

        // Build a lookup map for source ID -> Name
        const sourceMap: Record<number, string> = {};
        sourcesData.forEach(s => {
          sourceMap[s.ID] = s.Name;
        });

        setItems(feedData);
        setSources(sourceMap);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleInteract = async (id: number, action: 'save' | 'dismiss') => {
    // Optimistic UI update
    setHiddenItemIds(prev => new Set(prev).add(id));
    
    try {
      const res = await fetch(`/api/items/${id}/${action}`, { method: 'POST' });
      if (!res.ok) {
        throw new Error(`Failed to ${action} item`);
      }
    } catch (err) {
      console.error(err);
      // Revert optimistic update on failure
      setHiddenItemIds(prev => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
      alert(`Could not ${action} item. Please try again.`);
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Unknown date';
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    }).format(date);
  };

  // Filter out items the user has just interacted with
  const visibleItems = items.filter(item => !hiddenItemIds.has(item.ID));

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>myPersonalTechRadar</h1>
        <p>Curated signals from your trusted sources</p>
      </header>

      <main>
        {error && (
          <div className="error-state">
            <p><strong>Error:</strong> {error}</p>
          </div>
        )}

        {loading ? (
          <div className="loading">
            <div className="spinner"></div>
            <p>Discovering signals...</p>
          </div>
        ) : visibleItems.length === 0 ? (
          <div className="empty-state">
            <h3>You're all caught up!</h3>
            <p>No new items in your feed right now.</p>
          </div>
        ) : (
          <div className="feed-container">
            {visibleItems.map(item => (
              <article key={item.ID} className="feed-item">
                <div className="feed-item-content">
                  <h2 className="feed-item-title">
                    <a href={item.URL} target="_blank" rel="noopener noreferrer">
                      {item.Title}
                    </a>
                  </h2>
                  
                  <div className="feed-item-meta">
                    <span className="badge">{sources[item.SourceID] || 'Unknown Source'}</span>
                    <span>•</span>
                    <time dateTime={item.PublishedAt}>{formatDate(item.PublishedAt)}</time>
                    {item.Author && (
                      <>
                        <span>•</span>
                        <span>{item.Author}</span>
                      </>
                    )}
                  </div>
                  
                  {item.PersonalizationExplanation && (
                    <div style={{ fontSize: '0.8rem', color: 'var(--accent-success)', marginTop: '0.25rem', fontWeight: 500 }}>
                      ✓ {item.PersonalizationExplanation}
                    </div>
                  )}
                  
                  {item.Summary && (
                    <div className="feed-item-summary">
                      {stripHtml(item.Summary)}
                    </div>
                  )}

                  {item.WhyItMatters && (
                    <div style={{ marginTop: '0.75rem', padding: '0.75rem', backgroundColor: 'rgba(99, 102, 241, 0.1)', borderLeft: '3px solid var(--accent-primary)', borderRadius: '0 var(--radius-sm) var(--radius-sm) 0' }}>
                      <strong style={{ display: 'block', fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--accent-primary)', marginBottom: '0.25rem' }}>✨ Why It Matters</strong>
                      <span style={{ fontSize: '0.9rem', color: 'var(--text-primary)' }}>{item.WhyItMatters}</span>
                    </div>
                  )}
                </div>
                
                <div className="feed-item-actions">
                  <button 
                    className="btn btn-primary"
                    onClick={() => handleInteract(item.ID, 'save')}
                  >
                    Save for later
                  </button>
                  <button 
                    className="btn btn-outline-danger"
                    onClick={() => handleInteract(item.ID, 'dismiss')}
                  >
                    Dismiss
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
