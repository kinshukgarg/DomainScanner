import { useState } from 'react';
import './DomainScanner.css'; // Import CSS for styling

function DomainScanner() {
  const [domain, setDomain] = useState('');
  const [loading, setLoading] = useState(false);
  const [subdomains, setSubdomains] = useState([]);
  const [error, setError] = useState('');

  const handleInputChange = (e) => {
    setDomain(e.target.value);
  };

  const handleScan = async () => {
    if (!domain) {
      alert("Please enter a domain.");
      return;
    }

    setLoading(true);
    setSubdomains([]);
    setError('');

    try {
      const response = await fetch('http://localhost:8080/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain }),
      });

      if (!response.ok) {
        throw new Error('Failed to fetch scan results');
      }

      const data = await response.json();
      if (data.subdomains && data.subdomains.length > 0) {
        setSubdomains(data.subdomains);
      } else {
        setError("No subdomains or ports found.");
      }
    } catch (err) {
      setError('An error occurred: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="scanner-container">
      <h1>Domain Scanner</h1>
      <input
        type="text"
        value={domain}
        onChange={handleInputChange}
        placeholder="Enter domain (e.g., example.com)"
        className="domain-input"
      />
      <button onClick={handleScan} disabled={loading} className="scan-button">
        {loading ? "Scanning..." : "Start Scan"}
      </button>

      {error && <p className="error-message">{error}</p>}

      {subdomains.length > 0 && (
        <div className="subdomains-list">
          {subdomains.map((subdomain, index) => (
            <div key={index} className="subdomain-card">
              <h3>{subdomain.name}</h3>
              <p>Open Ports: {subdomain.ports.join(', ')}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default DomainScanner;
