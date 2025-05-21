# fullcycle-stresstest


docker build -t loadtester .
docker run loadtester --url=http://google.com --requests=100 --concurrency=10