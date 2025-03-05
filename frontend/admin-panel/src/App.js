import React, { useState, useEffect } from 'react'
import {
	BrowserRouter as Router,
	Route,
	Routes,
	Navigate,
} from 'react-router-dom'
import axios from 'axios'
import './App.css'

// Настройки API
const API_BASE_URL = 'http://0.0.0.0:3000' // Замените на ваш URL

// Интерсептор для добавления токена в заголовок Authorization
axios.interceptors.request.use(config => {
	const token = localStorage.getItem('token')
	if (token) {
		config.headers.Authorization = `Bearer ${token}`
	}
	return config
})

// Компонент логина
const LoginForm = ({ onLogin }) => {
	const [credentials, setCredentials] = useState({ email: '', password: '' })

	const handleSubmit = async e => {
		e.preventDefault()
		try {
			const response = await axios.post(`${API_BASE_URL}/login`, credentials)
			console.log(response.headers)
			localStorage.setItem('token', response.data.token) // Сохраняем токен
			onLogin() // Уведомляем App об успешной аутентификации
		} catch (error) {
			console.error('Login error:', error)
			alert('Login failed!')
		}
	}

	return (
		<div className='login-container'>
			<h2>Admin Login</h2>
			<form onSubmit={handleSubmit}>
				<div>
					<label>email:</label>
					<input
						type='text'
						value={credentials.email}
						onChange={e =>
							setCredentials({ ...credentials, email: e.target.value })
						}
					/>
				</div>
				<div>
					<label>Password:</label>
					<input
						type='password'
						value={credentials.password}
						onChange={e =>
							setCredentials({ ...credentials, password: e.target.value })
						}
					/>
				</div>
				<button type='submit'>Login</button>
			</form>
		</div>
	)
}

// Компонент добавления предметов
const AddItemForm = ({ onItemAdded }) => {
	const [newItem, setNewItem] = useState({ name: '', multiplier: 0, price: 0 })

	const handleSubmit = async e => {
		e.preventDefault()

		// Convert 'price' to int and 'multiplier' to float64 before sending
		const formattedItem = {
			name: newItem.name,
			multiplier: parseFloat(newItem.multiplier), // Ensure it's a float
			price: parseInt(newItem.price, 10), // Ensure it's an integer
		}

		try {
			await axios.post(`${API_BASE_URL}/api/items`, formattedItem)
			onItemAdded() // Update item list after adding
			setNewItem({ name: '', multiplier: 0, price: 0 }) // Reset form
		} catch (error) {
			console.error('Error adding item:', error)
			alert('Failed to add item!')
		}
	}

	return (
		<div className='item-form'>
			<h3>Add New Item</h3>
			<form onSubmit={handleSubmit}>
				<div>
					<label>Name:</label>
					<input
						type='text'
						value={newItem.name}
						onChange={e => setNewItem({ ...newItem, name: e.target.value })}
						required
					/>
				</div>
				<div>
					<label>Multiplier:</label>
					<input
						type='number'
						step='0.01'
						value={newItem.multiplier}
						onChange={e =>
							setNewItem({ ...newItem, multiplier: e.target.value })
						}
						required
					/>
				</div>
				<div>
					<label>Price:</label>
					<input
						type='number'
						step='0.01'
						value={newItem.price}
						onChange={e => setNewItem({ ...newItem, price: e.target.value })}
						required
					/>
				</div>
				<button type='submit'>Add Item</button>
			</form>
		</div>
	)
}

// Основная админ-панель
const AdminPanel = () => {
	const [items, setItems] = useState([])
	const [loading, setLoading] = useState(true)

	const fetchItems = async () => {
		try {
			const response = await axios.get(`${API_BASE_URL}/api/items`)
			setItems(response.data)
			setLoading(false)
		} catch (error) {
			console.error('Error fetching items:', error)
			setLoading(false)
		}
	}

	useEffect(() => {
		fetchItems()
	}, [])

	const handleLogout = () => {
		localStorage.removeItem('token') // Удаляем токен
		window.location.reload() // Перезагружаем страницу
	}

	return (
		<div className='admin-panel'>
			<div className='header'>
				<h2>Admin Panel</h2>
				<button onClick={handleLogout}>Logout</button>
			</div>

			<AddItemForm onItemAdded={fetchItems} />

			<h3>Items List</h3>
			{loading ? (
				<p>Loading items...</p>
			) : (
				<table>
					<thead>
						<tr>
							<th>Name</th>
							<th>Multiplier</th>
							<th>Price</th>
						</tr>
					</thead>
					<tbody>
						{items.map(item => (
							<tr key={item.id}>
								<td>{item.name}</td>
								<td>{item.multiplier}</td>
								<td>{item.price}</td>
							</tr>
						))}
					</tbody>
				</table>
			)}
		</div>
	)
}

// Главный компонент приложения
const App = () => {
	const [isAuthenticated, setIsAuthenticated] = useState(
		!!localStorage.getItem('token')
	)

	return (
		<Router>
			<div className='App'>
				<Routes>
					<Route
						path='/login'
						element={
							!isAuthenticated ? (
								<LoginForm onLogin={() => setIsAuthenticated(true)} />
							) : (
								<Navigate to='/' />
							)
						}
					/>
					<Route
						path='/'
						element={
							isAuthenticated ? <AdminPanel /> : <Navigate to='/login' />
						}
					/>
				</Routes>
			</div>
		</Router>
	)
}

export default App
