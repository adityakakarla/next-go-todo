'use client'

import { useEffect, useState } from 'react'
import { Loader2, CheckCircle2, Circle, Plus, Trash2 } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

interface Task {
  id: number
  title: string
  completed: boolean
}

export default function ApiDisplay() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [newTaskTitle, setNewTaskTitle] = useState('')

  const fetchTasks = async () => {
    try {
      const response = await fetch('http://localhost:8080/tasks')
      if (!response.ok) {
        throw new Error('Failed to fetch data')
      }
      const json: Task[] = await response.json()
      setTasks(json)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTasks()
  }, [])

  const handleAddTask = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTaskTitle.trim()) return

    try {
      const response = await fetch('http://localhost:8080/tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title: newTaskTitle }),
      })

      if (!response.ok) {
        throw new Error('Failed to add task')
      }

      setNewTaskTitle('')
      await fetchTasks() // Refresh the task list
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add task')
    }
  }

  const handleToggleTask = async (taskId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/tasks/toggle`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ id: taskId }),
      })

      if (!response.ok) {
        throw new Error('Failed to toggle task')
      }

      // Update the task in the local state
      setTasks(tasks.map(task => 
        task.id === taskId ? { ...task, completed: !task.completed } : task
      ))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to toggle task')
    }
  }

  const handleDeleteTask = async (taskId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/tasks/delete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ id: taskId }),
      })

      if (!response.ok) {
        throw new Error('Failed to delete task')
      }

      // Remove the task from the local state
      setTasks(tasks.filter(task => task.id !== taskId))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete task')
    }
  }

  return (
    <div className="min-h-screen dark flex items-start justify-center bg-black p-4">
      <Card className="w-full max-w-2xl bg-zinc-900 border-zinc-800 mt-24">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-zinc-100 text-xl font-medium">Tasks</CardTitle>
          <div className="text-sm text-zinc-500">localhost:8080</div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleAddTask} className="flex space-x-2 mb-4">
            <Input
              type="text"
              placeholder="Add a new task"
              value={newTaskTitle}
              onChange={(e) => setNewTaskTitle(e.target.value)}
              className="flex-grow bg-zinc-800 text-zinc-100 border-zinc-700"
            />
            <Button type="submit" className="bg-zinc-800 text-white hover:bg-zinc-900 transition ease-in-out duration-300">
              <Plus className="h-4 w-4 mr-2" />
              Add Task
            </Button>
          </form>

          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
            </div>
          ) : error ? (
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <p className="text-red-500">{error}</p>
            </div>
          ) : (
            <div className="space-y-4">
              {tasks.map((task) => (
                <div 
                  key={task.id} 
                  className="flex items-center justify-between p-4 rounded-lg bg-zinc-800"
                >
                  <div 
                    className="flex items-center space-x-4 cursor-pointer"
                    onClick={() => handleToggleTask(task.id)}
                  >
                    {task.completed ? (
                      <CheckCircle2 className="h-5 w-5 text-green-500" />
                    ) : (
                      <Circle className="h-5 w-5 text-zinc-500" />
                    )}
                    <span className={`text-zinc-100 ${task.completed ? 'line-through' : ''}`}>{task.title}</span>
                  </div>
                  <Button
                    onClick={() => handleDeleteTask(task.id)}
                    variant="ghost"
                    size="icon"
                    className="text-zinc-400 hover:text-red-500"
                  >
                    <Trash2 className="w-5 h-5" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

