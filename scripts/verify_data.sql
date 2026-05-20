SELECT COUNT(*) AS bar_count FROM bars WHERE symbol = 'AAPL';

SELECT MIN(time) AS earliest, MAX(time) AS latest
FROM bars WHERE symbol = 'AAPL';

SELECT time, open, high, low, close, volume
FROM bars WHERE symbol = 'AAPL'
ORDER BY time DESC
LIMIT 5;

SELECT COUNT(*) AS invalid_bars
FROM bars
WHERE high < low OR close > high OR close < low OR open > high OR open < low;
