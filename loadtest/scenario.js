import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
	vus: 10,
	duration: '1m',
};

const BASE_URL = 'http://localhost:8080';

export default function() {
	const teamPayload = JSON.stringify({
		team_name: 'backend',
		members: [
			{ user_id: 'u1', username: 'Alice', is_active: true },
			{ user_id: 'u2', username: 'Bob', is_active: true },
			{ user_id: 'u3', username: 'Eve', is_active: true },
		],
	});

	let res = http.post(`${BASE_URL}/team/add`, teamPayload, {
		headers: { 'Content-Type': 'application/json' },
	});
	check(res, { 'team.add is 201 or 400': (r) => r.status === 201 || r.status === 400 });

	const prId = `pr-${__VU}-${__ITER}`;
	const prPayload = JSON.stringify({
		pull_request_id: prId,
		pull_request_name: 'Test PR',
		author_id: 'u1',
	});
	res = http.post(`${BASE_URL}/pullRequest/create`, prPayload, {
		headers: { 'Content-Type': 'application/json' },
	});
	check(res, { 'pr.create is 201': (r) => r.status === 201 });

	const mergePayload = JSON.stringify({ pull_request_id: prId });
	res = http.post(`${BASE_URL}/pullRequest/merge`, mergePayload, {
		headers: { 'Content-Type': 'application/json' },
	});
	check(res, { 'pr.merge is 200': (r) => r.status === 200 });

	sleep(0.1);
}
